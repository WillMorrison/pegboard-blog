package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"time"

	"github.com/WillMorrison/pegboard-blog/grid"
	"github.com/WillMorrison/pegboard-blog/placer"
	"github.com/WillMorrison/pegboard-blog/sets"
	"github.com/WillMorrison/pegboard-blog/solver"
	"github.com/hashicorp/packer/command/enumflag"
)

const (
	UnorderedStonePlacer = "unordered"
	OrderedStonePlacer   = "ordered"

	EmptyStartingPoint         = "empty_grid"
	SingleOctantStartingPoints = "first_octant"

	MapSeparationSet = "map"
	BitSeparationSet = "array"
)

func main() {
	size := flag.Uint("size", 7, "the side length of square grid to search for solutions on")

	var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

	separationSet := BitSeparationSet
	flag.Var(enumflag.New(&separationSet, MapSeparationSet, BitSeparationSet), "separation_set", "SeparationSet implementation to use")

	stonePlacer := OrderedStonePlacer
	flag.Var(enumflag.New(&stonePlacer, UnorderedStonePlacer, OrderedStonePlacer), "placer", "StonePlacer implementation to use")

	startingPoint := SingleOctantStartingPoints
	flag.Var(enumflag.New(&startingPoint, EmptyStartingPoint, SingleOctantStartingPoints), "start", "Starting point for the search")

	flag.Parse()

	if *size > grid.MaxGridSize {
		log.Fatal("No solutions exist for 15x15 or larger grids. Not searching.")
	}
	g := grid.Grid{Size: uint8(*size)}

	var startingPointsProvider solver.StartingPointsProvider
	switch startingPoint {
	case EmptyStartingPoint:
		startingPointsProvider = solver.EmptyStartingPoint
	case SingleOctantStartingPoints:
		startingPointsProvider = solver.SingleOctantStartingPoints
	}

	var separationSetConstructor sets.SeparationSetConstructor
	switch separationSet {
	case MapSeparationSet:
		separationSetConstructor = sets.NewMapSeparationSet
	case BitSeparationSet:
		separationSetConstructor = sets.NewBitSeparationSet
	}

	var stonePlacerConstructor placer.StonePlacerConstructor
	switch stonePlacer {
	case UnorderedStonePlacer:
		stonePlacerConstructor = placer.UnorderedStonePlacerProvider{
			SeparationSetConstructor: separationSetConstructor,
			PointSetConstructor:      sets.NewMapPointSet}
	case OrderedStonePlacer:
		stonePlacerConstructor = placer.OrderedStonePlacerProvider{
			SeparationSetConstructor: separationSetConstructor}
	}

	s := solver.SingleThreadedSolver{
		StartingPointsProvider: startingPointsProvider,
		StonePlacerConstructor: stonePlacerConstructor,
	}

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	startTime := time.Now()
	solution, err := s.Solve(g)
	duration := time.Since(startTime)
	if err != nil {
		fmt.Printf("Search ended with no solution found for %+v in %v\n", g, duration)
		return
	}
	solution.Sort()
	if grid.IsValidSolution(g, solution) {
		fmt.Printf("Solution found for %+v in %v: %v\n", g, duration, solution)
	} else {
		fmt.Printf("We found a solution for %+v in %v but it was invalid! %v", g, duration, solution)
	}
}
