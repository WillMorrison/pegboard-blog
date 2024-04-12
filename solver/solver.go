package solver

import (
	"fmt"
	"runtime"
	"sync"

	"github.com/WillMorrison/pegboard-blog/grid"
	"github.com/WillMorrison/pegboard-blog/placer"
)

var (
	errNoSolutions = fmt.Errorf("no solutions exist")
)

type Solver interface {
	// Solve returns either Placements such that IsValidSolution(grid, placements) == true, or an error
	Solve(grid.Grid) (grid.Placements, error)
}

type StartingPointsProvider func(grid.Grid) []grid.Placements

// EmptyStartingPoint returns a single, empty Placements
func EmptyStartingPoint(g grid.Grid) []grid.Placements {
	return []grid.Placements{{}}
}

// SingleOctantStartingPoints returns Placements which have a single stone placed in the first octant going clockwise from the top left corner.
// All other starting positions can be made from reflections or rotations of these points, so we don't need to search them.
//
// Example: Single Octant starting points for a 5x5 grid, shown as *:
// * * * - -
// - * * - -
// - - * - -
// - - - - -
// - - - - -
func SingleOctantStartingPoints(g grid.Grid) []grid.Placements {
	var startingPoints []grid.Placements
	for i := uint8(0); i*2 < g.Size; i++ {
		for j := i; j*2 < g.Size; j++ {
			startingPoints = append(startingPoints, grid.Placements{grid.Point{Row: i, Col: j}})
		}
	}
	return startingPoints
}

type SingleThreadedSolver struct {
	StartingPointsProvider StartingPointsProvider
	StonePlacerConstructor placer.StonePlacerConstructor
}

func (s SingleThreadedSolver) dfs(sp placer.StonePlacer) (placer.StonePlacer, error) {
	if len(sp.Placements()) == int(sp.Grid().Size) {
		return sp, nil
	}

	for !sp.Done() {
		nextState, err := sp.Place()
		if err != nil {
			continue
		}
		final, err := s.dfs(nextState)
		if err != nil {
			continue
		}
		return final, nil
	}
	return sp, errNoSolutions
}

func (s SingleThreadedSolver) Solve(g grid.Grid) (grid.Placements, error) {
	for _, sp := range s.StartingPointsProvider(g) {
		start := s.StonePlacerConstructor.New(g, sp)
		solution, err := s.dfs(start)
		if err != nil {
			continue
		}
		return solution.Placements(), nil
	}
	return nil, errNoSolutions
}

type AsyncSolver struct {
	StartingPointsProvider StartingPointsProvider
	StonePlacerConstructor placer.StonePlacerConstructor
}

// dfs implements depth first search, and returns any found solutions on the solution channel.
// If the done channel is closed, the search is aborted
func (s AsyncSolver) dfs(sp placer.StonePlacer, solution chan<- grid.Placements, done <-chan struct{}) {
	for !sp.Done() {
		select {
		// If done channel is closed, abort search
		case <-done:
			return
		default:
		}
		nextState, err := sp.Place()
		if err != nil {
			continue
		}
		if len(nextState.Placements()) == int(nextState.Grid().Size) {
			solution <- nextState.Placements()
			return
		}
		s.dfs(nextState, solution, done)
	}
}

func (s AsyncSolver) Solve(g grid.Grid) (grid.Placements, error) {
	wg := sync.WaitGroup{}
	done := make(chan struct{})
	solutions := make(chan grid.Placements, 1)
	for _, sp := range s.StartingPointsProvider(g) {
		start := s.StonePlacerConstructor.New(g, sp)
		wg.Add(1)
		go func() {
			defer wg.Done()
			s.dfs(start, solutions, done)

		}()
	}
	go func() {
		// If wg.Wait returns, all dfs searches should have completed.
		wg.Wait()
		select {
		// They might have completed if one found a solution, in which case just abort
		case <-done:
			return
		// Or none might have found a solution, in which case send a nil to the solutions channel to unblock Solve's receiver
		// Keep in mind we might have returned from Wait before Solve closed done, so send nil in a nonblocking manner.
		case solutions <- nil:
		default:
		}
	}()

	solution := <-solutions
	close(done)
	if solution != nil {
		return solution, nil
	}
	return nil, errNoSolutions
}

type AsyncSplittingSolver struct {
	StartingPointsProvider StartingPointsProvider
	StonePlacerConstructor placer.StonePlacerConstructor
}

type workRequest struct {
	// The sender of the request owns the memory for the response placements, so provide that memory to the sender
	Placements grid.Placements
	// The channel that the requester will wait on for a response.
	Response   chan grid.Placements
}

// Send will reply to the request for work. It does not transfer ownership of the memory associated with the Placements slice.
// Returns when either the response is sent, or the done channel is closed.
func (wr *workRequest) Send(p grid.Placements, done <-chan struct{}) {
	wr.Placements = wr.Placements[:len(p)]
	copy(wr.Placements, p)
	select {
	case wr.Response <- wr.Placements:
	case <-done:
	}
}

// dfs implements depth first search, and returns any found solutions on the solution channel.
// If the done channel is closed, the search is aborted
// Work is split as requests are available in the work channel
func (s AsyncSplittingSolver) dfs(sp placer.StonePlacer, solution chan<- grid.Placements, done <-chan struct{}, work chan *workRequest) {
	for !sp.Done() {
		select {
		// If done channel is closed, abort search
		case <-done:
			return
		default:
		}
		nextState, err := sp.Place()
		if err != nil {
			continue
		}
		if len(nextState.Placements()) == int(nextState.Grid().Size) {
			solution <- nextState.Placements()
			return
		}

		select {
		// Split work if there is a request in the work channel. The requesting worker will eventually pick up this part of the search and we can move on.
		case request := <-work:
			request.Send(nextState.Placements(), done)
		default:
			s.dfs(nextState, solution, done, work)
		}
	}
}

// worker adds requests to the work channel when idle, and listens for tasks to come back or the done channel to be closed.
func (s AsyncSplittingSolver) worker(g grid.Grid, solutions chan<- grid.Placements, done <-chan struct{}, work chan *workRequest) {
	request := workRequest{
		Placements: make(grid.Placements, 0, g.Size),
		Response:   make(chan grid.Placements),
	}
	for {
		select {
		case work <- &request: // Request some work to do
			select {
			case p := <-request.Response:
				sp := s.StonePlacerConstructor.New(g, p)
				s.dfs(sp, solutions, done, work)
			case <-done:
				return
			}
		case <-done: // Exit if a solution was found by some worker
			return
		}
	}
}

func (s AsyncSplittingSolver) Solve(g grid.Grid) (grid.Placements, error) {
	numWorkers := runtime.NumCPU()

	wg := sync.WaitGroup{}
	work := make(chan *workRequest, numWorkers)
	done := make(chan struct{})
	solutions := make(chan grid.Placements, 1)

	// Add starting points to work queue
	wg.Add(1)
	go func() {
		defer wg.Done()
		for _, sp := range s.StartingPointsProvider(g) {
			select {
			case request := <-work:
				request.Send(sp, done) // Queue some work to do
			case <-done: // Exit if a solution was found by some worker
				return
			}
		}
	}()

	// Start workers
	for i := 0; i < numWorkers; i++ {
		go func() {
			s.worker(g, solutions, done, work)
		}()
	}

	go func() {
		// If wg.Wait returns, initial load should have finished.
		wg.Wait()
		// Wait for all workers to be waiting on requests
		for len(work) != numWorkers {
			select {
			// They might have completed if one found a solution, in which case just abort
			case <-done:
				return
			default:
			}
		}
		select {
		// They might have completed if one found a solution, in which case just abort
		case <-done:
			return
		// Or none might have found a solution, in which case send a nil to the solutions channel to unblock Solve's receiver
		// Keep in mind we might have returned from Wait before Solve closed done, so send nil in a nonblocking manner.
		case solutions <- nil:
		default:
		}
	}()

	solution := <-solutions
	close(done)
	if solution != nil {
		return solution, nil
	}
	return nil, errNoSolutions
}
