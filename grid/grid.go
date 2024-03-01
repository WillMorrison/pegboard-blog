// Package grid defines basic types and utility functions on those types
package grid

import (
	"fmt"
)

// Grids larger than 14x14 are known to have no solutions
const MaxGridSize = 14

// Grid represents an NxN square grid
type Grid struct {
	Size uint8
}

// Point is the coordinate of a stone on a grid
type Point struct {
	Row uint8
	Col uint8
}

func (p Point) String() string {
	return string('A'+p.Row) + fmt.Sprint(p.Col)
}

// IsInBounds returns whether a Point is contained within a given Grid
func IsInBounds(g Grid, p Point) bool {
	return p.Row < g.Size && p.Col < g.Size
}

func LessThan(p1, p2 Point) bool {
	return p1.Row < p2.Row || p1.Row == p2.Row && p1.Col < p2.Col
}

// Placements represents a set of stones placed on the grid
type Placements []Point

// Separation is the squared distance between 2 grid points
func Separation(p1, p2 Point) uint16 {
	return uint16((int16(p1.Row)-int16(p2.Row))*(int16(p1.Row)-int16(p2.Row)) + (int16(p1.Col)-int16(p2.Col))*(int16(p1.Col)-int16(p2.Col)))
}

// Checks that a proposed solution to the problem is valid
func IsValidSolution(g Grid, p Placements) bool {
	// Check that the required number of stones have been placed
	if len(p) != int(g.Size) {
		return false
	}

	separations := make(map[uint16]bool)
	for i, p1 := range p {
		// Check that all stones are in bounds
		if !IsInBounds(g, p1) {
			return false
		}

		for j := i + 1; j < len(p); j++ {
			p2 := p[j]
			s := Separation(p1, p2)
			// Check that no two stones are placed on the same point
			if s == 0 {
				return false
			}
			// Check that all separations are unique
			if separations[s] {
				return false
			}
			separations[s] = true
		}
	}

	return true
}
