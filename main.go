package main

import (
	"fmt"
	"time"

	"github.com/WillMorrison/pegboard-blog/grid"
	"github.com/WillMorrison/pegboard-blog/solver"
)

func main() {
	g := grid.Grid{Size: 8}
	s := solver.NewSingleThreadedSolver(g, solver.EmptyStartingPoint)
	startTime := time.Now()
	solution, err := s.Solve(g)
	duration := time.Since(startTime)
	if err != nil {
		fmt.Printf("Search ended with no solution found for %+v in %v\n", g, duration)
		return
	}
	if grid.IsValidSolution(g, solution) {
		fmt.Printf("Solution found for %+v in %v: %v\n", g, duration, solution)
	} else {
		fmt.Printf("We found a solution for %+v in %v but it was invalid! %v", g, duration, solution)
	}
}
