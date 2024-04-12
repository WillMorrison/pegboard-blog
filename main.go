package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime/pprof"
	"runtime/trace"
	"time"

	"github.com/WillMorrison/pegboard-blog/grid"
	"github.com/WillMorrison/pegboard-blog/placer"
	"github.com/WillMorrison/pegboard-blog/pruner"
	"github.com/WillMorrison/pegboard-blog/sets"
	"github.com/WillMorrison/pegboard-blog/solver"
	"github.com/hashicorp/packer/command/enumflag"
)

const (
	UnorderedStonePlacer                          = "unordered"
	OrderedStonePlacer                            = "ordered"
	OrderedNoAllocStonePlacer                     = "ordered_noalloc"
	OrderedNoAllocPruningStonePlacer              = "ordered_noalloc_pruning"
	OrderedNoAllocOpportunisticPruningStonePlacer = "ordered_noalloc_opportunistic_pruning"

	EmptyStartingPoint         = "empty_grid"
	SingleOctantStartingPoints = "first_octant"

	MapSeparationSet = "map"
	BitSeparationSet = "array"

	RuntimePruner     = "runtime"
	PrecomputedPruner = "precomputed"

	SingleThreadedSolver = "single_thread"
	AsyncSolver          = "async"
	AsyncSplittingSolver = "async_splitting"
)

func main() {
	size := flag.Uint("size", 7, "the side length of square grid to search for solutions on")

	var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")
	var memprofile = flag.String("memprofile", "", "write memory profile to this file")
	var tracefile = flag.String("trace", "", "write trace to this file")

	separationSet := BitSeparationSet
	flag.Var(enumflag.New(&separationSet, MapSeparationSet, BitSeparationSet), "separation_set", "SeparationSet implementation to use")

	prunerImpl := PrecomputedPruner
	flag.Var(enumflag.New(&prunerImpl, RuntimePruner, PrecomputedPruner), "pruner", "Pruner implementation to use")

	stonePlacer := OrderedNoAllocStonePlacer
	flag.Var(enumflag.New(&stonePlacer, UnorderedStonePlacer, OrderedStonePlacer, OrderedNoAllocStonePlacer, OrderedNoAllocPruningStonePlacer, OrderedNoAllocOpportunisticPruningStonePlacer), "placer", "StonePlacer implementation to use")

	startingPoint := SingleOctantStartingPoints
	flag.Var(enumflag.New(&startingPoint, EmptyStartingPoint, SingleOctantStartingPoints), "start", "Starting point for the search")

	solverImpl := AsyncSolver
	flag.Var(enumflag.New(&solverImpl, SingleThreadedSolver, AsyncSolver, AsyncSplittingSolver), "solver", "Solver implementation to use")

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
		separationSetConstructor = sets.NewBitArraySeparationSet
	}

	var prunerConstructor func(grid.Grid) pruner.Pruner
	switch prunerImpl {
	case RuntimePruner:
		prunerConstructor = pruner.NewRuntimePruner
	case PrecomputedPruner:
		prunerConstructor = pruner.NewPrecomputedPruner
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
	case OrderedNoAllocStonePlacer:
		stonePlacerConstructor = placer.OrderedNoAllocStonePlacerProvider{}
	case OrderedNoAllocPruningStonePlacer:
		stonePlacerConstructor = placer.OrderedPruningNoAllocStonePlacerProvider{
			PrunerConstructor: prunerConstructor,
		}
	case OrderedNoAllocOpportunisticPruningStonePlacer:
		stonePlacerConstructor = placer.OrderedOpportunisticPruningNoAllocStonePlacerProvider{
			PrunerConstructor: prunerConstructor,
		}
	}

	var s solver.Solver
	switch solverImpl {
	case SingleThreadedSolver:
		s = solver.SingleThreadedSolver{
			StartingPointsProvider: startingPointsProvider,
			StonePlacerConstructor: stonePlacerConstructor,
		}
	case AsyncSolver:
		s = solver.AsyncSolver{
			StartingPointsProvider: startingPointsProvider,
			StonePlacerConstructor: stonePlacerConstructor,
		}
	case AsyncSplittingSolver:
		s = solver.AsyncSplittingSolver{
			StartingPointsProvider: startingPointsProvider,
			StonePlacerConstructor: stonePlacerConstructor,
		}
	}

	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if *tracefile != "" {
		f, err := os.Create(*tracefile)
		if err != nil {
			log.Fatal(err)
		}
		trace.Start(f)
		defer trace.Stop()
	}

	startTime := time.Now()
	solution, err := s.Solve(g)
	duration := time.Since(startTime)

	if *memprofile != "" {
		f, err := os.Create(*memprofile)
		defer f.Close()
		if err != nil {
			log.Fatal(err)
		}
		err = pprof.WriteHeapProfile(f)
		if err != nil {
			log.Fatal(err)
		}
	}

	if err != nil {
		fmt.Printf("Search ended with no solution found for %+v in %v\n", g, duration)
		return
	}
	solution.Sort()
	if err := grid.CheckValidSolution(g, solution); err == nil {
		fmt.Printf("Solution found for %+v in %v: %v\n", g, duration, solution)
	} else {
		fmt.Printf("We found a solution %v for %+v in %v but it was invalid! %s\n", solution, g, duration, err)
	}
}
