package solver

import (
	"fmt"

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

func (sts SingleThreadedSolver) Solve(g grid.Grid) (grid.Placements, error) {
	for _, sp := range sts.StartingPointsProvider(g) {
		start := sts.StonePlacerConstructor.New(g, sp)
		solution, err := sts.dfs(start)
		if err != nil {
			continue
		}
		return solution.Placements(), nil
	}
	return nil, errNoSolutions
}
