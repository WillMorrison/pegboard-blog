package placer

import (
	"fmt"

	"github.com/WillMorrison/pegboard-blog/grid"
	"github.com/WillMorrison/pegboard-blog/pruner"
	"github.com/WillMorrison/pegboard-blog/sets"
)

var (
	errDistanceConstraintViolated = fmt.Errorf("cannot place stone, unique distance constraint would be violated")
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

// orderedStonePlacer attempts to place stones from top to bottom, left to right, checking that they are valid placements each time.
type orderedStonePlacer struct {
	grid        grid.Grid
	stones      grid.Placements
	separations sets.SeparationSet
	nextStone   grid.Point
}

func (sp *orderedStonePlacer) Place() (StonePlacer, error) {
	defer func() { sp.nextStone = grid.AdvanceStone(sp.grid, sp.nextStone) }()

	// Check that placing the next stone doesn't result in duplicate separations
	separations := sp.separations.Copy()
	for _, p := range sp.stones {
		s := grid.Separation(sp.nextStone, p)
		if separations.Has(s) {
			return sp, errDistanceConstraintViolated
		}
		separations.Add(s)
	}

	// Add the stone to a fresh copy of the placements slice
	newPlacements := make(grid.Placements, len(sp.stones))
	copy(newPlacements, sp.stones)
	newPlacements = append(newPlacements, sp.nextStone)

	return &orderedStonePlacer{sp.grid, newPlacements, separations, grid.AdvanceStone(sp.grid, sp.nextStone)}, nil
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
		nextStone = grid.AdvanceStone(g, p[len(p)-1])
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
	sp.nextStone = grid.AdvanceStone(sp.grid, sp.nextStone)
	for sp.stones.Has(sp.nextStone) {
		sp.nextStone = grid.AdvanceStone(sp.grid, sp.nextStone)
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
			return sp, errDistanceConstraintViolated
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

type orderedNoAllocStonePlacer struct {
	grid        grid.Grid
	stones      grid.Placements
	separations sets.SeparationSet
	nextStone   grid.Point
	nextPlacer  *orderedNoAllocStonePlacer
}

func (sp *orderedNoAllocStonePlacer) Place() (StonePlacer, error) {
	defer func() { sp.nextStone = grid.AdvanceStone(sp.grid, sp.nextStone) }()

	// Check that placing the next stone doesn't result in duplicate separations
	sp.nextPlacer.separations.Clone(sp.separations)
	for _, p := range sp.stones {
		s := grid.Separation(sp.nextStone, p)
		if sp.nextPlacer.separations.Has(s) {
			return nil, errDistanceConstraintViolated
		}
		sp.nextPlacer.separations.Add(s)
	}

	copy(sp.nextPlacer.stones, sp.stones)
	sp.nextPlacer.stones[len(sp.stones)] = sp.nextStone
	sp.nextPlacer.nextStone = grid.AdvanceStone(sp.grid, sp.nextStone)
	return sp.nextPlacer, nil
}

func (sp orderedNoAllocStonePlacer) Done() bool {
	return !grid.IsInBounds(sp.grid, sp.nextStone)
}

func (sp orderedNoAllocStonePlacer) Grid() grid.Grid {
	return sp.grid
}

func (sp orderedNoAllocStonePlacer) Placements() grid.Placements {
	return sp.stones
}

type OrderedNoAllocStonePlacerProvider struct{}

func (spp OrderedNoAllocStonePlacerProvider) New(g grid.Grid, p grid.Placements) StonePlacer {
	// Create a singly linked list of placers. the first will have 0 stones placed, the second 1 stone placed, and so on.
	placers := make([]orderedNoAllocStonePlacer, g.Size+1)
	for i := 0; i < len(placers); i++ {
		placers[i] = orderedNoAllocStonePlacer{
			grid:        g,
			stones:      make(grid.Placements, i),
			separations: sets.NewBitArraySeparationSet(nil), // This implementation's Clone() shouldn't allocate
			nextStone:   grid.Point{},
		}
		if i+1 < len(placers) {
			placers[i].nextPlacer = &(placers[i+1])
		}
	}
	// Place the stones, in order.
	p.Sort()
	for i, stone := range p {
		placers[i].nextStone = stone
		placers[i].Place()
	}
	// Return the placer with all the starting stones placed.
	return &placers[len(p)]
}

type orderedPruningNoAllocStonePlacer struct {
	grid        grid.Grid
	stones      grid.Placements
	separations sets.BitArraySeparationSet
	pruner      pruner.Pruner
	pruned      sets.BitArrayPointSet
	nextStone   grid.Point
	nextPlacer  *orderedPruningNoAllocStonePlacer
}

// Advance moves nextStone to the next non-pruned position, or leaves it out of bounds
func (sp *orderedPruningNoAllocStonePlacer) advance() {
	for sp.nextStone = grid.AdvanceStone(sp.grid, sp.nextStone); grid.IsInBounds(sp.grid, sp.nextStone); sp.nextStone = grid.AdvanceStone(sp.grid, sp.nextStone) {
		if !sp.pruned.Has(sp.nextStone) {
			return
		}
	}
}

func (sp *orderedPruningNoAllocStonePlacer) Place() (StonePlacer, error) {
	defer sp.advance()

	sp.nextPlacer.separations.Clone(&sp.separations)
	sp.nextPlacer.pruned.Clone(&sp.pruned)

	// prune isoceles triangles between nextStone and all previous stones.
	var newSeparations [grid.MaxGridSize]uint16 // track newly added separations apart from existing ones
	for i, p := range sp.stones {
		s := grid.Separation(sp.nextStone, p)
		if sp.nextPlacer.separations.Has(s) {
			return nil, errDistanceConstraintViolated
		}
		sp.nextPlacer.separations.Add(s)
		newSeparations[i] = s
		sp.nextPlacer.pruner.PruneIsoceles(&sp.nextPlacer.pruned, p, sp.nextStone)
	}

	// prune circles around existing points with new separations
	for i := 0; i < len(sp.stones); i++ {
		for _, p := range sp.stones {
			sp.nextPlacer.pruner.PruneCircles(&sp.nextPlacer.pruned, p, newSeparations[i])
		}
	}

	// prune circles around nextStone with existing+new separations
	allSepIter := sets.NewSeparationSetIteratorForGrid(&sp.nextPlacer.separations, sp.grid)
	for sep, ok := allSepIter.Next(); ok; sep, ok = allSepIter.Next() {
		sp.nextPlacer.pruner.PruneCircles(&sp.nextPlacer.pruned, sp.nextStone, sep)
	}

	// Add stone to placements
	copy(sp.nextPlacer.stones, sp.stones)
	sp.nextPlacer.stones[len(sp.stones)] = sp.nextStone

	sp.nextPlacer.nextStone = sp.nextStone
	sp.nextPlacer.advance()
	return sp.nextPlacer, nil
}

func (sp orderedPruningNoAllocStonePlacer) Done() bool {
	return !grid.IsInBounds(sp.grid, sp.nextStone)
}

func (sp orderedPruningNoAllocStonePlacer) Grid() grid.Grid {
	return sp.grid
}

func (sp orderedPruningNoAllocStonePlacer) Placements() grid.Placements {
	return sp.stones
}

type OrderedPruningNoAllocStonePlacerProvider struct {
	PrunerConstructor func(grid.Grid) pruner.Pruner
}

func (spp OrderedPruningNoAllocStonePlacerProvider) New(g grid.Grid, p grid.Placements) StonePlacer {
	pruner := spp.PrunerConstructor(g)

	// Create a singly linked list of placers. the first will have 0 stones placed, the second 1 stone placed, and so on.
	placers := make([]orderedPruningNoAllocStonePlacer, g.Size+1)
	for i := 0; i < len(placers); i++ {
		placers[i] = orderedPruningNoAllocStonePlacer{
			grid:        g,
			stones:      make(grid.Placements, i),
			separations: sets.BitArraySeparationSet{},
			pruner:      pruner,
			pruned:      sets.BitArrayPointSet{},
			nextStone:   grid.Point{},
		}
		if i+1 < len(placers) {
			placers[i].nextPlacer = &(placers[i+1])
		}
	}
	// Place the stones, in order.
	p.Sort()
	for i, stone := range p {
		if placers[i].pruned.Has(stone) {
			panic("Invalid placement, already pruned")
		}
		placers[i].nextStone = stone
		placers[i].Place()
	}
	// Return the placer with all the starting stones placed.
	return &placers[len(p)]
}

type orderedOpportunisticPruningNoAllocStonePlacer struct {
	grid        grid.Grid
	stones      grid.Placements
	separations sets.BitArraySeparationSet
	pruner      pruner.Pruner
	pruned      sets.BitArrayPointSet
	nextStone   grid.Point
	nextPlacer  *orderedOpportunisticPruningNoAllocStonePlacer
}

func (sp *orderedOpportunisticPruningNoAllocStonePlacer) advance() {
	for sp.nextStone = grid.AdvanceStone(sp.grid, sp.nextStone); grid.IsInBounds(sp.grid, sp.nextStone); sp.nextStone = grid.AdvanceStone(sp.grid, sp.nextStone) {
		if !sp.pruned.Has(sp.nextStone) {
			return
		}
	}
}

func (sp *orderedOpportunisticPruningNoAllocStonePlacer) Place() (StonePlacer, error) {
	defer sp.advance()

	sp.nextPlacer.separations.Clone(&sp.separations)
	sp.nextPlacer.pruned.Clone(&sp.pruned)

	// prune isoceles triangles between nextStone and all previous stones.
	for _, p := range sp.stones {
		s := grid.Separation(sp.nextStone, p)
		if sp.nextPlacer.separations.Has(s) {
			return nil, errDistanceConstraintViolated
		}
		sp.nextPlacer.separations.Add(s)
		sp.nextPlacer.pruner.PruneIsoceles(&sp.nextPlacer.pruned, p, sp.nextStone)
		sp.nextPlacer.pruner.PruneCircles(&sp.nextPlacer.pruned, p, s)
		sp.nextPlacer.pruner.PruneCircles(&sp.nextPlacer.pruned, sp.nextStone, s)
	}

	// Add stone to placements
	copy(sp.nextPlacer.stones, sp.stones)
	sp.nextPlacer.stones[len(sp.stones)] = sp.nextStone

	sp.nextPlacer.nextStone = sp.nextStone
	sp.nextPlacer.advance()
	return sp.nextPlacer, nil
}

func (sp orderedOpportunisticPruningNoAllocStonePlacer) Done() bool {
	return !grid.IsInBounds(sp.grid, sp.nextStone)
}

func (sp orderedOpportunisticPruningNoAllocStonePlacer) Grid() grid.Grid {
	return sp.grid
}

func (sp orderedOpportunisticPruningNoAllocStonePlacer) Placements() grid.Placements {
	return sp.stones
}

type OrderedOpportunisticPruningNoAllocStonePlacerProvider struct {
	PrunerConstructor func(grid.Grid) pruner.Pruner
}

func (spp OrderedOpportunisticPruningNoAllocStonePlacerProvider) New(g grid.Grid, p grid.Placements) StonePlacer {
	pruner := spp.PrunerConstructor(g)

	// Create a singly linked list of placers. the first will have 0 stones placed, the second 1 stone placed, and so on.
	placers := make([]orderedOpportunisticPruningNoAllocStonePlacer, g.Size+1)
	for i := 0; i < len(placers); i++ {
		placers[i] = orderedOpportunisticPruningNoAllocStonePlacer{

			grid:        g,
			stones:      make(grid.Placements, i),
			separations: sets.BitArraySeparationSet{},
			pruner:      pruner,
			pruned:      sets.BitArrayPointSet{},
			nextStone:   grid.Point{},
		}
		if i+1 < len(placers) {
			placers[i].nextPlacer = &(placers[i+1])
		}
	}
	// Place the stones, in order.
	p.Sort()
	for i, stone := range p {
		if placers[i].pruned.Has(stone) {
			panic("Invalid placement, already pruned")
		}
		placers[i].nextStone = stone
		placers[i].Place()
	}
	// Return the placer with all the starting stones placed.
	return &placers[len(p)]
}
