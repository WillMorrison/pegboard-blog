package placer

import (
	"fmt"

	"github.com/WillMorrison/pegboard-blog/grid"
	"github.com/WillMorrison/pegboard-blog/sets"
)

type StonePlacer interface {
	// Place attempts to place a stone. If placement is successful, it returns a new StonePlacer, otherwise it returns an error.
	Place() (StonePlacer, error)

	// Done returns whether any more placements are possible.
	Done() bool

	// Grid returns the Grid onto which stones are being placed
	Grid() grid.Grid

	// Placements returns the placements made so far.
	Placements() grid.Placements
}

type StonePlacerConstructor interface {
	// New returns a new StonePlacer that places on the given grid, with the given existing stones.
	New(grid.Grid, grid.Placements) StonePlacer
}

// advanceStone returns the next point in an ordered left to right, top to bottom traversal of the grid. The returned point is *not* guaranteed to be on the grid.
func advanceStone(g grid.Grid, p grid.Point) grid.Point {
	p2 := grid.Point{Row: p.Row, Col: p.Col + 1}
	if p2.Col == g.Size {
		p2 = grid.Point{Row: p.Row + 1, Col: 0}
	}
	return p2
}

// orderedStonePlacer attempts to place stones from top to bottom, left to right, checking that they are valid placements each time.
type orderedStonePlacer struct {
	grid        grid.Grid
	stones      grid.Placements
	separations sets.SeparationSet
	nextStone   grid.Point
}

func (sp *orderedStonePlacer) Place() (StonePlacer, error) {
	defer func() { sp.nextStone = advanceStone(sp.grid, sp.nextStone) }()

	// Check that placing the next stone doesn't result in duplicate separations
	separations := sp.separations.Copy()
	for _, p := range sp.stones {
		s := grid.Separation(sp.nextStone, p)
		if separations.Has(s) {
			return sp, fmt.Errorf("cannot place at %s, unique distance constraint violated with stone at %s", sp.nextStone, p)
		}
		separations.Add(s)
	}

	// Add the stone to a fresh copy of the placements slice
	newPlacements := make(grid.Placements, len(sp.stones))
	copy(newPlacements, sp.stones)
	newPlacements = append(newPlacements, sp.nextStone)

	return &orderedStonePlacer{sp.grid, newPlacements, separations, advanceStone(sp.grid, sp.nextStone)}, nil
}

func (sp orderedStonePlacer) Done() bool {
	return !grid.IsInBounds(sp.grid, sp.nextStone)
}

func (sp orderedStonePlacer) Grid() grid.Grid {
	return sp.grid
}

func (sp orderedStonePlacer) Placements() grid.Placements {
	return sp.stones
}

type OrderedStonePlacerProvider struct {
	SeparationSetConstructor sets.SeparationSetConstructor
}

func (spp OrderedStonePlacerProvider) New(g grid.Grid, p grid.Placements) StonePlacer {
	nextStone := grid.Point{}
	if len(p) > 0 {
		nextStone = advanceStone(g, p[len(p)-1])
	}
	return &orderedStonePlacer{grid: g, stones: p, separations: spp.SeparationSetConstructor(p), nextStone: nextStone}
}

// unorderedStonePlacer places stones in any unoccupied spot on the board
type unorderedStonePlacer struct {
	grid        grid.Grid
	stones      sets.PointSet
	separations sets.SeparationSet
	nextStone   grid.Point
}

// advance moves nextStone to a point that is not already occupied
func (sp *unorderedStonePlacer) advance() {
	sp.nextStone = advanceStone(sp.grid, sp.nextStone)
	for sp.stones.Has(sp.nextStone) {
		sp.nextStone = advanceStone(sp.grid, sp.nextStone)
	}
}

func (sp *unorderedStonePlacer) Place() (StonePlacer, error) {
	if sp.stones.Has(sp.nextStone) {
		sp.advance()
	}
	defer sp.advance()

	// Check that placing the next stone doesn't result in duplicate separations
	separations := sp.separations.Copy()
	for _, p := range sp.stones.Elements() {
		s := grid.Separation(sp.nextStone, p)
		if separations.Has(s) {
			return sp, fmt.Errorf("cannot place at %s, unique distance constraint violated with stone at %s", sp.nextStone, p)
		}
		separations.Add(s)
	}

	// Add the stone to a fresh copy of the placements
	newStones := sp.stones.Copy()
	newStones.Add(sp.nextStone)

	return &unorderedStonePlacer{sp.grid, newStones, separations, grid.Point{}}, nil
}

func (sp unorderedStonePlacer) Done() bool {
	return !grid.IsInBounds(sp.grid, sp.nextStone)
}

func (sp unorderedStonePlacer) Grid() grid.Grid {
	return sp.grid
}

func (sp unorderedStonePlacer) Placements() grid.Placements {
	return sp.stones.Elements()
}

type UnorderedStonePlacerProvider struct {
	SeparationSetConstructor sets.SeparationSetConstructor
	PointSetConstructor      sets.PointSetConstructor
}

func (spp UnorderedStonePlacerProvider) New(g grid.Grid, p grid.Placements) StonePlacer {
	return &unorderedStonePlacer{grid: g, stones: spp.PointSetConstructor(p), separations: spp.SeparationSetConstructor(p), nextStone: grid.Point{}}
}
