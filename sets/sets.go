package sets

import (
	"unsafe"

	"github.com/WillMorrison/pegboard-blog/grid"
)

type SeparationSet interface {
	Has(uint16) bool
	Add(uint16)
	Union(SeparationSet)
	Clear()
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

func (ss mapSeparationSet) Union(ss2 SeparationSet) {
	for _, sep := range ss2.Elements() {
		ss[sep] = true
	}
}

func (ss mapSeparationSet) Clear() {
	for k := range ss {
		delete(ss, k)
	}
}

func (ss mapSeparationSet) Copy() SeparationSet {
	newSet := make(mapSeparationSet)
	for s := range ss {
		newSet[s] = true
	}
	return newSet
}

func (ss mapSeparationSet) Clone(ss2 SeparationSet) {
	ss.Clear()
	ss.Union(ss2)
}

func (ss mapSeparationSet) Elements() []uint16 {
	keys := make([]uint16, 0, len(ss))
	for k := range ss {
		keys = append(keys, k)
	}
	return keys
}

// A set representing membership as bits. Has up to 2*13^2 = 338 members, which is sufficient for separations on a max sized grid.
// Separation element ordering is little endian.
type BitArraySeparationSet [6]uint64

func NewBitArraySeparationSet(p grid.Placements) SeparationSet {
	var s BitArraySeparationSet
	for i, p1 := range p {
		for j := i + 1; j < len(p); j++ {
			p2 := p[j]
			s.Add(grid.Separation(p1, p2))
		}
	}
	return &s
}

func (ss BitArraySeparationSet) Has(sep uint16) bool {
	return ss[sep>>6]&(0x1<<(sep&0x3f)) != 0
}

func (ss *BitArraySeparationSet) Add(sep uint16) {
	ss[sep>>6] |= 0x1 << (sep & 0x3f)
}

func (ss *BitArraySeparationSet) Union(ss2 SeparationSet) {
	switch t := ss2.(type) {
	// If the second set is also a bit array, just bitwise or the array
	case *BitArraySeparationSet:
		// unrolled loop for speed
		ss[0] |= t[0]
		ss[1] |= t[1]
		ss[2] |= t[2]
		ss[3] |= t[3]
		ss[4] |= t[4]
		ss[5] |= t[5]
	default:
		for _, sep := range ss2.Elements() {
			ss.Add(sep)
		}
	}

}

func (ss *BitArraySeparationSet) Clear() {
	*ss = BitArraySeparationSet{}
}

func (ss *BitArraySeparationSet) Copy() SeparationSet {
	var newSet BitArraySeparationSet = *ss
	return &newSet
}

func (ss *BitArraySeparationSet) Clone(ss2 SeparationSet) {
	switch t := ss2.(type) {
	// If the second set is also a bit array, just copy the array
	case *BitArraySeparationSet:
		*ss = *t
	default:
		ss.Clear()
		ss.Union(ss2)
	}
}

func (ss BitArraySeparationSet) Elements() []uint16 {
	keys := make([]uint16, 0, len(ss))
	for sep := uint16(0); sep < uint16(grid.MaxSeparation+1); sep++ {
		if ss.Has(sep) {
			keys = append(keys, sep)
		}
	}
	return keys
}

type SeparationSetIterator struct {
	SeparationSet SeparationSet
	sep           uint16
	maxSep        uint16
}

func NewSeparationSetIterator(ss SeparationSet) SeparationSetIterator {
	ssi := SeparationSetIterator{SeparationSet: ss, maxSep: grid.MaxSeparation}
	for ssi.sep++; ssi.sep < ssi.maxSep+1 && !ssi.SeparationSet.Has(ssi.sep); ssi.sep++ {
	}
	return ssi
}

func NewSeparationSetIteratorForGrid(ss SeparationSet, g grid.Grid) SeparationSetIterator {
	ssi := SeparationSetIterator{SeparationSet: ss, maxSep: uint16(g.Size-1) * uint16(g.Size-1) * 2}
	for ssi.sep++; ssi.sep < ssi.maxSep+1 && !ssi.SeparationSet.Has(ssi.sep); ssi.sep++ {
	}
	return ssi
}

func (ssi *SeparationSetIterator) Next() (uint16, bool) {
	if ssi.sep > ssi.maxSep {
		return 0, false
	}
	ret := ssi.sep
	for ssi.sep++; ssi.sep < ssi.maxSep+1 && !ssi.SeparationSet.Has(ssi.sep); ssi.sep++ {
	}
	return ret, true
}

type PointSet interface {
	// Has checks if the point is in the set
	Has(grid.Point) bool
	// Add adds the point to the set
	Add(grid.Point)
	// Union updates the set to contain the union of points of the two sets
	Union(PointSet)
	// Clear resets the set to contain no points
	Clear()
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
	for p, ok := it.Next(); ok; p, ok = it.Next() {
		ps1.Add(p)
	}
}
func genericPointSetClone(ps1, ps2 PointSet) {
	ps1.Clear()
	genericPointSetUnion(ps1, ps2)
}

type placementsIterator struct {
	i        int
	elements grid.Placements
}

func (pi *placementsIterator) Next() (grid.Point, bool) {
	if pi.i == len(pi.elements) {
		return grid.Point{}, false
	}
	next := pi.elements[pi.i]
	pi.i++
	return next, true
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

func (ps mapPointSet) Clear() {
	for k := range ps {
		delete(ps, k)
	}
}

func (ps mapPointSet) Copy() PointSet {
	newSet := make(mapPointSet)
	for p := range ps {
		newSet[p] = true
	}
	return newSet
}

func (ps mapPointSet) Clone(ps2 PointSet) {
	genericPointSetClone(ps, ps2)
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

// A set representing membership as bits. Has up to 16^2 = 256 members, which is sufficient for all points on a max sized grid.
// Each uint16 represents memberships for one row.
type BitArrayPointSet [16]uint16

func NewBitArrayPointSet(points grid.Placements) PointSet {
	var ps BitArrayPointSet
	for _, p := range points {
		ps.Add(p)
	}
	return &ps
}

type bitArrayPointSetIterator struct {
	ps   *BitArrayPointSet
	next grid.Point
}

func (pi *bitArrayPointSetIterator) Next() (grid.Point, bool) {
	if pi.next.Row >= grid.MaxGridSize {
		return pi.next, false
	}
	next := pi.next
	for pi.next = grid.AdvanceStone(grid.Grid{grid.MaxGridSize}, pi.next); pi.next.Row < grid.MaxGridSize; pi.next = grid.AdvanceStone(grid.Grid{grid.MaxGridSize}, pi.next) {
		// Skip over empty rows without iterating through columns
		for pi.next.Row < grid.MaxGridSize && pi.ps[pi.next.Row] == 0 {
			pi.next.Row++
			pi.next.Col = 0
		}
		if pi.ps.Has(pi.next) {
			break
		}
	}
	return next, true
}

func (ps BitArrayPointSet) Has(p grid.Point) bool {
	return ps[p.Row]&(0x8000>>p.Col) != 0
}

func (ps *BitArrayPointSet) Add(p grid.Point) {
	ps[p.Row] |= 0x8000 >> p.Col
}

func (ps *BitArrayPointSet) Union(ps2 PointSet) {
	switch t := ps2.(type) {
	// If the second set is also a bit array, use bitwise or
	case *BitArrayPointSet:
		// unrolled loop and typecasting for speed
		v1 := (*[4]uint64)(unsafe.Pointer(ps))
		v2 := (*[4]uint64)(unsafe.Pointer(t))
		v1[0] |= v2[0]
		v1[1] |= v2[1]
		v1[2] |= v2[2]
		v1[3] |= v2[3]
		// for i := 0; i < len(ps); i++ {
		// 	ps[i] |= t[i]
		// }
	default:
		genericPointSetUnion(ps, ps2)
	}
}

func (ps *BitArrayPointSet) Clear() {
	*ps = BitArrayPointSet{}
}

func (ps *BitArrayPointSet) Copy() PointSet {
	var newSet BitArrayPointSet = *ps
	return &newSet
}

func (ps *BitArrayPointSet) Clone(ps2 PointSet) {
	switch t := ps2.(type) {
	// If the second set is also a bit array, just copy the array
	case *BitArrayPointSet:
		*ps = *t
	default:
		genericPointSetClone(ps, ps2)
	}
}

func (ps BitArrayPointSet) Elements() grid.Placements {
	keys := make(grid.Placements, 0, len(ps))
	it := ps.Iter()
	for p, ok := it.Next(); ok; p, ok = it.Next() {
		keys = append(keys, p)
	}
	return keys
}

func (ps *BitArrayPointSet) Iter() grid.PointIterator {
	it := bitArrayPointSetIterator{ps: ps, next: grid.Point{}}
	if !ps.Has(it.next) {
		it.Next()
	}
	return &it
}
