package solver

import (
	"fmt"

	"github.com/WillMorrison/pegboard-blog/grid"
)

type StartingPointsProvider func(grid.Grid) []grid.Placements

// EmptyStartingPoint returns a single, empty Placements
func EmptyStartingPoint(g grid.Grid) []grid.Placements {
	return []grid.Placements{{}}
}

// SingleOctantStartingPoints returns Placements which have a single stone placed in the first octant going clockwise from the top left corner.
// All other starting positions can be made from reflections or rotations of these points, so we don't need to search them.
func SingleOctantStartingPoints(g grid.Grid) []grid.Placements {
	var startingPoints []grid.Placements
	for i := 0; i*2 < int(g.Size); i++ {
		for j := i; j*2 < int(g.Size); j++ {
			startingPoints = append(startingPoints, grid.Placements{grid.Point{Row: uint8(i), Col: uint8(j)}})
		}
	}
	return startingPoints
}

type SeparationSet interface {
	In(uint16) bool
	Add(uint16)
	Copy() SeparationSet
	String() string
}

type SeparationSetConstructor func(grid.Placements) SeparationSet

// a map-based set for keeping track of separation distances
type mapSeparationSet map[uint16]bool

func NewMapSeparationSet(p grid.Placements) SeparationSet {
	s := make(mapSeparationSet)
	for i, p1 := range p {
		for j := i + 1; j < len(p); j++ {
			p2 := p[j]
			s[grid.Separation(p1, p2)] = true
		}
	}
	return s
}

func (mss mapSeparationSet) In(sep uint16) bool {
	return mss[sep]
}

func (mss mapSeparationSet) Add(sep uint16) {
	mss[sep] = true
}

func (mss mapSeparationSet) Copy() SeparationSet {
	newSet := make(mapSeparationSet)
	for s := range mss {
		newSet[s] = true
	}
	return newSet
}

func (mss mapSeparationSet) String() string {
	keys := make([]uint16, 0, len(mss))
	for k := range mss {
		keys = append(keys, k)
	}
	return fmt.Sprint(keys)

}

type partialSolution struct {
	g           grid.Grid
	stones      grid.Placements
	separations SeparationSet
}

// TryPlace attempts to place a stone at point p, and returns the resulting partial solution.
// It is an error to place a stone such that it is on another stone, to the left of or above an existing stone,
// or such that it results in two pairs of stones being the same distance apart.
func (ps partialSolution) TryPlace(p grid.Point) (partialSolution, error) {
	// Check stone placement ordering restriction
	if len(ps.stones) > 0 {
		lastStone := ps.stones[len(ps.stones)-1]
		if (lastStone.Row == p.Row && p.Col <= lastStone.Col) || p.Row < lastStone.Row {
			return partialSolution{}, fmt.Errorf("cannot place at %s, ordering constraint violated with stone at %s", p, lastStone)
		}
	}

	// Create a copy, we're about to mutate the separations set
	separations := ps.separations.Copy()

	// check that placing the stone doesn't result in duplicate separations
	for _, p1 := range ps.stones {
		s := grid.Separation(p1, p)
		if separations.In(s) {
			return partialSolution{}, fmt.Errorf("cannot place at %s, unique distance constraint violated with stone at %s", p, p1)
		}
		separations.Add(s)
	}

	newPlacements := make(grid.Placements, len(ps.stones))
	copy(newPlacements, ps.stones)
	newPlacements = append(newPlacements, p)
	return partialSolution{ps.g, newPlacements, separations}, nil
}

func (ps partialSolution) Done() bool {
	return len(ps.stones) == int(ps.g.Size)
}

type SingleThreadedSolver struct {
	grid                     grid.Grid
	startingPointsProvider   StartingPointsProvider
	separationSetConstructor SeparationSetConstructor
}

func (sts SingleThreadedSolver) dfs(ps partialSolution) (partialSolution, error) {
	if ps.Done() {
		return ps, nil
	}

	var nextPoint grid.Point
	if len(ps.stones) > 0 {
		nextPoint = ps.stones[len(ps.stones)-1]
		nextPoint.Col++
	}
	for ; nextPoint.Row < ps.g.Size; nextPoint.Row++ {
		for ; nextPoint.Col < ps.g.Size; nextPoint.Col++ {
			newSolution, err := ps.TryPlace(nextPoint)
			if err != nil {
				continue
			}
			final, err := sts.dfs(newSolution)
			if err != nil {
				continue
			}
			return final, nil
		}
		nextPoint.Col = 0
	}
	return ps, fmt.Errorf("no solutions found")
}

func (sts SingleThreadedSolver) Solve(g grid.Grid) (grid.Placements, error) {
	for _, sp := range sts.startingPointsProvider(sts.grid) {
		ps := partialSolution{g, sp, sts.separationSetConstructor(sp)}
		solution, err := sts.dfs(ps)
		if err != nil {
			continue
		}
		return solution.stones, nil
	}
	return nil, fmt.Errorf("no solutions exist")
}

func NewSingleThreadedSolver(g grid.Grid, spp StartingPointsProvider) Solver {
	return SingleThreadedSolver{g, spp, NewMapSeparationSet}
}
