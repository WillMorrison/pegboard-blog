package sets

import (
	"github.com/WillMorrison/pegboard-blog/grid"
)

type SeparationSet interface {
	Has(uint16) bool
	Add(uint16)
	Copy() SeparationSet
	Clone(SeparationSet)
	Elements() []uint16
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

func (ss mapSeparationSet) Clone(ss2 SeparationSet) {
	for k := range ss {
		delete(ss, k)
	}
	for _, sep := range ss2.Elements() {
		ss[sep] = true
	}
}

func (ss mapSeparationSet) Elements() []uint16 {
	keys := make([]uint16, 0, len(ss))
	for k := range ss {
		keys = append(keys, k)
	}
	return keys
}

// A set representing membership as bits. Has up to 2*14^2 = 392 members, which is sufficient for separations on a max sized grid.
// Separation element ordering is little endian across the whole array.
type bitSeparationSet [49]byte

func NewBitSeparationSet(p grid.Placements) SeparationSet {
	var s bitSeparationSet
	for i, p1 := range p {
		for j := i + 1; j < len(p); j++ {
			p2 := p[j]
			s.Add(grid.Separation(p1, p2))
		}
	}
	return &s
}

func (ss bitSeparationSet) Has(sep uint16) bool {
	return ss[sep>>3]&(0x80>>(sep&0x7)) != 0
}

func (ss *bitSeparationSet) Add(sep uint16) {
	ss[sep>>3] |= 0x80 >> (sep & 0x7)
}

func (ss *bitSeparationSet) Copy() SeparationSet {
	var newSet bitSeparationSet = *ss
	return &newSet
}

func (ss *bitSeparationSet) Clone(ss2 SeparationSet) {
	switch t := ss2.(type) {
	// If the second set is also a bit array, just copy the array
	case *bitSeparationSet:
		*ss = *t
	default:
		for i := 0; i < len(ss); i++ {
			ss[i] = 0
		}
		for _, sep := range ss2.Elements() {
			ss.Add(sep)
		}
	}
}

func (ss bitSeparationSet) Elements() []uint16 {
	keys := make([]uint16, 0, len(ss))
	for sep := uint16(0); sep < uint16(len(ss)*8); sep++ {
		if ss.Has(sep) {
			keys = append(keys, sep)
		}
	}
	return keys
}

type PointSet interface {
	// Has checks if the point is in the set
	Has(grid.Point) bool
	// Add adds the point to the set
	Add(grid.Point)
	// Union updates the set to contain the union of points of the two sets
	Union(PointSet)
	// Copy creates a copy of the set that does not share memory
	Copy() PointSet
	// Clone updates the set to contain the same elements as the other set
	Clone(PointSet)
	// Elements returns a slice of points in the set
	Elements() grid.Placements
	// Iter returns an iterator over the points in the set
	Iter() grid.PointIterator
}

type PointSetConstructor func(grid.Placements) PointSet

func genericPointSetUnion(ps1, ps2 PointSet) {
	it := ps2.Iter()
	for p, done := it.Next(); done == nil; p, done = it.Next() {
		ps1.Add(p)
	}
}

type placementsIterator struct {
	i        int
	elements grid.Placements
}

func (pi *placementsIterator) Next() (grid.Point, error) {
	if pi.i == len(pi.elements) {
		return grid.Point{}, grid.ErrIterationFinished
	}
	next := pi.elements[pi.i]
	pi.i++
	return next, nil
}

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

func (ps mapPointSet) Union(ps2 PointSet) {
	genericPointSetUnion(ps, ps2)
}

func (ps mapPointSet) Copy() PointSet {
	newSet := make(mapPointSet)
	for p := range ps {
		newSet[p] = true
	}
	return newSet
}

func (ps mapPointSet) Clone(ps2 PointSet) {
	for k := range ps {
		delete(ps, k)
	}
	genericPointSetUnion(ps, ps2)
}

func (ps mapPointSet) Elements() grid.Placements {
	points := make(grid.Placements, 0, len(ps))
	for p := range ps {
		points = append(points, p)
	}
	return points
}

func (ps mapPointSet) Iter() grid.PointIterator {
	return &placementsIterator{i: 0, elements: ps.Elements()}
}

type bitArrayPointSet [16]uint16

func NewBitArrayPointSet(points grid.Placements) PointSet {
	var ps bitArrayPointSet
	for _, p := range points {
		ps.Add(p)
	}
	return &ps
}

type bitArrayPointSetIterator struct {
	ps   *bitArrayPointSet
	next grid.Point
}

func (pi *bitArrayPointSetIterator) Next() (grid.Point, error) {
	if pi.next.Row >= grid.MaxGridSize {
		return pi.next, grid.ErrIterationFinished
	}
	next := pi.next
	for pi.next = grid.AdvanceStone(grid.Grid{grid.MaxGridSize}, pi.next); pi.next.Row < grid.MaxGridSize; pi.next = grid.AdvanceStone(grid.Grid{grid.MaxGridSize}, pi.next) {
		// Skip over empty rows without iterating through columns
		for pi.next.Row < grid.MaxGridSize && pi.ps[pi.next.Row] == 0 {
			pi.next.Row++
			pi.next.Col = 0
		}
		if pi.ps.Has(pi.next) {
			return next, nil
		}
	}
	return next, nil
}

func (ps bitArrayPointSet) Has(p grid.Point) bool {
	return ps[p.Row]&(0x8000>>p.Col) != 0
}

func (ps *bitArrayPointSet) Add(p grid.Point) {
	ps[p.Row] |= 0x8000 >> p.Col
}

func (ps *bitArrayPointSet) Union(ps2 PointSet) {
	switch t := ps2.(type) {
	// If the second set is also a bit array, use bitwise or
	case *bitArrayPointSet:
		for i := 0; i < len(ps); i++ {
			ps[i] |= t[i]
		}
	default:
		genericPointSetUnion(ps, ps2)
	}
}

func (ps *bitArrayPointSet) Copy() PointSet {
	var newSet bitArrayPointSet = *ps
	return &newSet
}

func (ps *bitArrayPointSet) Clone(ps2 PointSet) {
	switch t := ps2.(type) {
	// If the second set is also a bit array, just copy the array
	case *bitArrayPointSet:
		*ps = *t
	default:
		*ps = bitArrayPointSet{}
		genericPointSetUnion(ps, ps2)
	}
}

func (ps bitArrayPointSet) Elements() grid.Placements {
	keys := make(grid.Placements, 0, len(ps))
	it := ps.Iter()
	for p, done := it.Next(); done == nil; p, done = it.Next() {
		keys = append(keys, p)
	}
	return keys
}

func (ps *bitArrayPointSet) Iter() grid.PointIterator {
	it := bitArrayPointSetIterator{ps: ps, next: grid.Point{}}
	if !ps.Has(it.next) {
		it.Next()
	}
	return &it
}
