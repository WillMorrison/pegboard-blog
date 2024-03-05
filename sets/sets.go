package sets

import (
	"fmt"

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
	Has(grid.Point) bool
	Add(grid.Point)
	Copy() PointSet
	Clone(PointSet)
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

func (ss mapPointSet) Clone(ss2 PointSet) {
	for k := range ss {
		delete(ss, k)
	}
	for _, sep := range ss2.Elements() {
		ss[sep] = true
	}
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

type bitArrayPointSet [32]byte

func NewBitArrayPointSet(points grid.Placements) PointSet {
	var ps bitArrayPointSet
	for _, p := range points {
		ps.Add(p)
	}
	return &ps
}

// pointToIndex encodes a point into one byte, storing the row in the first 4 bits, and the column in the last 4
func pointToIndex(p grid.Point) uint8 {
	return p.Row<<4 | p.Col
}

// indexToPoint decodes a Point from a byte created by pointToIndex.
func indexToPoint(i uint8) grid.Point {
	return grid.Point{Row: i >> 4, Col: i & 0x0f}
}

func (ps *bitArrayPointSet) Has(p grid.Point) bool {
	i := pointToIndex(p)
	return ps[i>>3]&(0x80>>(i&0x7)) != 0
}

func (ps *bitArrayPointSet) Add(p grid.Point) {
	i := pointToIndex(p)
	ps[i>>3] |= 0x80 >> (i & 0x7)
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
		for i := 0; i < len(ps); i++ {
			ps[i] = 0
		}
		for _, sep := range ps2.Elements() {
			ps.Add(sep)
		}
	}
}

func (ps bitArrayPointSet) Elements() grid.Placements {
	keys := make(grid.Placements, 0, len(ps))
	for i := 0; i < len(ps); i++ {
		for j := uint8(0); j < 8; j++ {
			if ps[i]&(0x80>>j) != 0 {
				index := uint8(i)<<3 + j
				keys = append(keys, indexToPoint(index))
			}
		}
	}
	return keys
}
