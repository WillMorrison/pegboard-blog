package solver

import "github.com/WillMorrison/pegboard-blog/grid"

type Solver interface {
	// Solve returns either Placements such that IsValidSolution(grid, placements) == true, or an error
	Solve(grid.Grid) (grid.Placements, error)
}