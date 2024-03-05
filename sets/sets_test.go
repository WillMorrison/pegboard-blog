package sets

import (
	"testing"

	"github.com/WillMorrison/pegboard-blog/grid"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

func Test_SeparationSet(t *testing.T) {
	// the largest possible separation on a max sized grid .
	// Some implementations have an upper bound on separation that is not 2^16
	maxSep := grid.Separation(grid.Point{0, 0}, grid.Point{grid.MaxGridSize - 1, grid.MaxGridSize - 1})

	tests := []struct {
		name string
		ssc  SeparationSetConstructor
	}{
		{"mapSeparationSet", NewMapSeparationSet},
		{"bitSeparationSet", NewBitSeparationSet},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Run("Empty_Has_Elements", func(t *testing.T) {
				ss := tt.ssc(nil)
				if ss.Has(0) {
					t.Errorf("%s.Has(0)=true, want false", tt.name)
				}
				if got := len(ss.Elements()); got != 0 {
					t.Errorf("len(%s.Elements())=%d, want 0", tt.name, got)
				}
			})

			t.Run("Add_Has_Elements", func(t *testing.T) {
				for sep := uint16(0); sep <= maxSep; sep++ {
					ss := tt.ssc(nil)
					ss.Add(sep)
					if !ss.Has(sep) {
						t.Errorf("%s.Has(%d)=false, want true", tt.name, sep)
					}
					if got := len(ss.Elements()); got != 1 {
						t.Errorf("len(%s.Elements())=%d, want 1", tt.name, got)
					}
				}
			})

			t.Run("Add_Has_Other", func(t *testing.T) {
				for sep := uint16(0); sep < maxSep; sep++ {
					ss := tt.ssc(nil)
					ss.Add(sep)
					if ss.Has(sep + 1) {
						t.Errorf("%s.Has(%d)=true, want false", tt.name, sep+1)
					}
				}
			})

			t.Run("Add_Copy_Has_Elements", func(t *testing.T) {
				sep := uint16(4)
				ss1 := tt.ssc(nil)
				ss1.Add(sep)
				ss2 := ss1.Copy()
				if !ss2.Has(sep) {
					t.Errorf("%s.Has(%d)=false, want true", tt.name, sep)
				}
				if got := len(ss2.Elements()); got != 1 {
					t.Errorf("len(%s.Elements())=%d, want 1", tt.name, got)
				}
			})

			t.Run("Copy_Add_Has_Elements", func(t *testing.T) {
				sep := uint16(4)
				ss1 := tt.ssc(nil)
				ss2 := ss1.Copy()
				ss2.Add(sep)
				if got := len(ss1.Elements()); got != 0 {
					t.Errorf("len(%s.Elements())=%d, want 0", tt.name, got)
				}
			})

			t.Run("Constructor", func(t *testing.T) {
				ss := tt.ssc(grid.Placements{ // Separation matrix for points noted in comments
					grid.Point{0, 0}, //  0  25  25   5
					grid.Point{3, 4}, // 25   0  10   8
					grid.Point{0, 5}, // 25  10   0  10
					grid.Point{1, 2}, //  5   8  10   0
				})
				want := []uint16{5, 8, 10, 25}
				if got := ss.Elements(); !cmp.Equal(got, want, cmpopts.SortSlices(func(a, b uint16) bool { return a < b })) {
					t.Errorf("len(%s.Elements())=%d, want 0", tt.name, got)
				}
			})

			t.Run("Add_Clone_Elements", func(t *testing.T) {
				// Add two different separations to each set, then make the second set a clone of the first
				sep1 := uint16(4)
				sep2 := uint16(6)
				ss1 := tt.ssc(nil)
				ss1.Add(sep1)
				ss2 := tt.ssc(nil)
				ss2.Add(sep2)
				ss2.Clone(ss1)
				if diff := cmp.Diff(ss1.Elements(), ss2.Elements()); diff != "" {
					t.Errorf("%s.Clone().Elements() had diff %s", tt.name, diff)
				}
			})

			t.Run("Clone_Add_Has", func(t *testing.T) {
				// Make the second set a clone of the first, then add a value to it
				sep := uint16(4)
				ss1 := tt.ssc(nil)
				ss2 := tt.ssc(nil)
				ss2.Clone(ss1)
				ss2.Add(sep)
				if ss1.Has(sep) {
					t.Errorf("%s.Has(%d)=true, want false", tt.name, sep)
				}
			})
		})
	}
}

func Test_bitSeparationSet_Clone_mapSeparationSet(t *testing.T) {
	sep1 := uint16(4)
	sep2 := uint16(6)
	ss1 := NewMapSeparationSet(nil)
	ss1.Add(sep1)
	ss2 := NewBitSeparationSet(nil)
	ss2.Add(sep2)
	ss2.Clone(ss1)
	if diff := cmp.Diff(ss1.Elements(), ss2.Elements()); diff != "" {
		t.Errorf("bitSeparationset.Clone(mapSeparationSet).Elements() had diff %s", diff)
	}
}

func Test_PointSet(t *testing.T) {
	// Arbitrary grid point values.
	point1 := grid.Point{Row: 1, Col: 2}
	point2 := grid.Point{Row: 3, Col: 4}

	tests := []struct {
		name string
		psc  PointSetConstructor
	}{
		{"mapPointSet", NewMapPointSet},
		{"bitArrayPointSet", NewBitArrayPointSet},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			t.Run("Empty Has", func(t *testing.T) {
				ps := tt.psc(nil)
				if ps.Has(point1) {
					t.Errorf("%s.Has(%s)=true, want false", ps, point1)
				}
			})

			t.Run("Add Has", func(t *testing.T) {
				ps := tt.psc(nil)
				ps.Add(point1)
				if !ps.Has(point1) {
					t.Errorf("%s.Has(%s)=false, want true", ps, point1)
				}
			})

			t.Run("Add Has Other", func(t *testing.T) {
				ps := tt.psc(nil)
				ps.Add(point1)
				if ps.Has(point2) {
					t.Errorf("%s.Has(%s)=true, want false", ps, point2)
				}
			})

			t.Run("Add Copy Has", func(t *testing.T) {
				ps1 := tt.psc(nil)
				ps1.Add(point1)
				ps2 := ps1.Copy()
				if !ps2.Has(point1) {
					t.Errorf("%s.Has(%s)=false, want true", ps2, point1)
				}
			})

			t.Run("Add Copy Has Other", func(t *testing.T) {
				ps1 := tt.psc(nil)
				ps1.Add(point1)
				ps2 := ps1.Copy()
				if ps2.Has(point2) {
					t.Errorf("%s.Has(%s)=true, want false", ps2, point2)
				}
			})

			t.Run("Copy Add Has", func(t *testing.T) {
				ps1 := tt.psc(nil)
				ps2 := ps1.Copy()
				ps2.Add(point1)
				if ps1.Has(point1) {
					t.Errorf("%s.Has(%s)=true, want false", ps1, point1)
				}
			})

			t.Run("Elements", func(t *testing.T) {
				tests := []struct {
					name string
					arg  grid.Placements
					want grid.Placements
				}{
					{
						name: "nil",
						arg:  nil,
						want: grid.Placements{},
					},
					{
						name: "empty",
						arg:  grid.Placements{},
						want: grid.Placements{},
					},
					{
						name: "nonempty",
						arg:  grid.Placements{point1, point2},
						want: grid.Placements{point1, point2},
					},
				}
				for _, ttt := range tests {
					t.Run(ttt.name, func(t *testing.T) {
						ps := tt.psc(ttt.arg)
						if got := ps.Elements(); !cmp.Equal(got, ttt.want, cmpopts.SortSlices(grid.LessThan)) {
							t.Errorf("%s(%v).Elements() = %v, want %v", tt.name, ttt.arg, got, ttt.want)
						}
					})
				}
			})
			t.Run("Add_Clone_Elements", func(t *testing.T) {
				// Add two different points to each set, then make the second set a clone of the first
				ps1 := tt.psc(nil)
				ps1.Add(point1)
				ps2 := tt.psc(nil)
				ps2.Add(point2)
				ps2.Clone(ps1)
				if diff := cmp.Diff(ps1.Elements(), ps2.Elements()); diff != "" {
					t.Errorf("%s.Clone().Elements() had diff %s", tt.name, diff)
				}
			})

			t.Run("Clone_Add_Has", func(t *testing.T) {
				// Make the second set a clone of the first, then add a value to it
				ps1 := tt.psc(nil)
				ps2 := tt.psc(nil)
				ps2.Clone(ps1)
				ps2.Add(point1)
				if ps1.Has(point1) {
					t.Errorf("%s.Has(%d)=true, want false", tt.name, point1)
				}
			})

		})
	}
}

func Test_bitArrayPointSet_Clone_mapPointSet(t *testing.T) {
	// Arbitrary grid point values.
	point1 := grid.Point{Row: 1, Col: 2}
	point2 := grid.Point{Row: 3, Col: 4}
	ps1 := NewMapPointSet(nil)
	ps1.Add(point1)
	ps2 := NewBitArrayPointSet(nil)
	ps2.Add(point2)
	ps2.Clone(ps1)
	if diff := cmp.Diff(ps1.Elements(), ps2.Elements()); diff != "" {
		t.Errorf("bitArrayPointSet.Clone(mapPointSet).Elements() had diff %s", diff)
	}
}

func Test_bitArrayPointSet_MaxGridPoints(t *testing.T) {
	ps := NewBitArrayPointSet(nil)
	for row := uint8(0); row < grid.MaxGridSize; row++ {
		for col := uint8(0); col < grid.MaxGridSize; col++ {
			p1 := grid.Point{Row: row, Col: col}
			ps.Add(p1)

			p2 := indexToPoint(pointToIndex(p1))
			if p1 != p2 {
				t.Errorf("indexToPoint(pointToIndex(%v))=%v, want %v", p1, p2, p1)
			}
		}
	}
	want := grid.MaxGridSize * grid.MaxGridSize
	if got := len(ps.Elements()); got != want {
		t.Errorf("Pointset has %d elements, want %d", got, want)
	}
}
