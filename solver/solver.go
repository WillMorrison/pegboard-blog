package solver

import (
	"fmt"
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
