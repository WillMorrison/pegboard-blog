package pruner

import (
	"sync"

	"github.com/WillMorrison/pegboard-blog/grid"
	"github.com/WillMorrison/pegboard-blog/sets"
)

type Pruner interface {
	// PruneIsoceles updates the given set to include all points that form an isoceles triangle with the two given points
	PruneIsoceles(sets.PointSet, grid.Point, grid.Point)
	// PruneCircles updates the given set to include all points that fall on the circle with the given radius (squared) around the given point
	PruneCircles(sets.PointSet, grid.Point, uint16)
}

type runtimePruner struct {
	grid grid.Grid
}

func NewRuntimePruner(g grid.Grid) Pruner {
	return runtimePruner{grid: g}
}

func (p runtimePruner) PruneIsoceles(ps sets.PointSet, p1, p2 grid.Point) {
	// This implementation is rather inefficient because it iterates over the whole grid.
	// We could do better, but this Pruner will soon be replaced by a cached precomputation which only runs this once
	it := p.grid.Iter()
	for p3, ok := it.Next(); ok; p3, ok = it.Next() {
		if grid.Separation(p1, p3) == grid.Separation(p2, p3) {
			ps.Add(p3)
		}
	}
}

func (p runtimePruner) PruneCircles(ps sets.PointSet, p1 grid.Point, sep uint16) {
	// This implementation is rather inefficient because it iterates over the whole grid.
	// We could do better, but this Pruner will soon be replaced by a cached precomputation which only runs this once
	it := p.grid.Iter()
	for p2, ok := it.Next(); ok; p2, ok = it.Next() {
		if grid.Separation(p1, p2) == sep {
			ps.Add(p2)
		}
	}
}

type precomputedPruner struct {
	isoceles [grid.MaxGridSize][grid.MaxGridSize][grid.MaxGridSize][grid.MaxGridSize]sets.BitArrayPointSet
	circles  [grid.MaxGridSize][grid.MaxGridSize][grid.MaxSeparation + 1]sets.BitArrayPointSet
}

// Global singleton instances of precomputedPruner by grid size
var (
	mu                       sync.Mutex
	cachedPrecomputedPruners []*precomputedPruner = make([]*precomputedPruner, grid.MaxGridSize)
)

func NewPrecomputedPruner(g grid.Grid) Pruner {
	mu.Lock()
	defer mu.Unlock()
	if pruner := cachedPrecomputedPruners[g.Size-1]; pruner != nil {
		return pruner
	}
	rp := runtimePruner{g}
	p := new(precomputedPruner)
	it1 := g.Iter()
	for p1, ok1 := it1.Next(); ok1; p1, ok1 = it1.Next() {
		it2 := g.Iter()
		for p2, ok2 := it2.Next(); ok2; p2, ok2 = it2.Next() {
			if p1 == p2 {
				continue
			}
			sep := grid.Separation(p1, p2)
			rp.PruneCircles(&(p.circles[p1.Row][p1.Col][sep]), p1, sep)
			rp.PruneIsoceles(&(p.isoceles[p1.Row][p1.Col][p2.Row][p2.Col]), p1, p2)
		}
	}
	cachedPrecomputedPruners[g.Size-1] = p
	return p
}

func (p *precomputedPruner) PruneIsoceles(ps sets.PointSet, p1, p2 grid.Point) {
	ps.Union(&p.isoceles[p1.Row][p1.Col][p2.Row][p2.Col])
}

func (p *precomputedPruner) PruneCircles(ps sets.PointSet, p1 grid.Point, sep uint16) {
	ps.Union(&p.circles[p1.Row][p1.Col][sep])
}
