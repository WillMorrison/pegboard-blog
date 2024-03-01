package main

import (
	"flag"
	"fmt"
	"time"

	"github.com/WillMorrison/pegboard-blog/grid"
	"github.com/WillMorrison/pegboard-blog/placer"
	"github.com/WillMorrison/pegboard-blog/sets"
	"github.com/WillMorrison/pegboard-blog/solver"
)

func main() {
	size := flag.Uint("size", 7, "the side length of square grid to search for solutions on")

	flag.Parse()

	if *size > grid.MaxGridSize {
		fmt.Println("No solutions exist for 15x15 or larger grids. Not searching.")
		return
	}
	g := grid.Grid{Size: uint8(*size)}

	s := solver.SingleThreadedSolver{
		StartingPointsProvider: solver.SingleOctantStartingPoints,
		StonePlacerConstructor: placer.OrderedStonePlacerProvider{sets.NewMapSeparationSet},
	}
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
