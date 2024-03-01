package sets

import (
	"fmt"

	"github.com/WillMorrison/pegboard-blog/grid"
)

type SeparationSet interface {
	Has(uint16) bool
	Add(uint16)
	Copy() SeparationSet
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

func (ss mapSeparationSet) Has(sep uint16) bool {
	return ss[sep]
}

func (ss mapSeparationSet) Add(sep uint16) {
	ss[sep] = true
}

func (ss mapSeparationSet) Copy() SeparationSet {
	newSet := make(mapSeparationSet)
	for s := range ss {
		newSet[s] = true
	}
	return newSet
}

func (ss mapSeparationSet) String() string {
	keys := make([]uint16, 0, len(ss))
	for k := range ss {
		keys = append(keys, k)
	}
	return fmt.Sprint(keys)

}

type PointSet interface {
	Has(grid.Point) bool
	Add(grid.Point)
	Copy() PointSet
	Elements() grid.Placements
}

type PointSetConstructor func(grid.Placements) PointSet

type mapPointSet map[grid.Point]bool

func NewMapPointSet(points grid.Placements) PointSet {
	ps := make(mapPointSet)
	for _, p := range points {
		ps[p] = true
	}
	return ps
}

func (ps mapPointSet) Has(p grid.Point) bool {
	return ps[p]
}

func (ps mapPointSet) Add(p grid.Point) {
	ps[p] = true
}

func (ps mapPointSet) Copy() PointSet {
	newSet := make(mapPointSet)
	for p := range ps {
		newSet[p] = true
	}
	return newSet
}

func (ps mapPointSet) Elements() grid.Placements {
	points := make(grid.Placements, 0, len(ps))
	for p := range ps {
		points = append(points, p)
	}
	return points
}

func (ps mapPointSet) String() string {
	return fmt.Sprint(ps.Elements())

}
