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

func (mss mapSeparationSet) Has(sep uint16) bool {
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
