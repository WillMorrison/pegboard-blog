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
		{"bitSeparationSet", NewBitArraySeparationSet},
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

			t.Run("Add_Clear_Has", func(t *testing.T) {
				ss := tt.ssc(nil)
				ss.Add(maxSep)
				ss.Clear()
				if got := len(ss.Elements()); got != 0 {
					t.Errorf("len(%s.Clear().Elements())=%d, want 0", tt.name, got)
				}
			})

			t.Run("Union_Elements", func(t *testing.T) {
				ss1 := tt.ssc(nil)
				ss1.Add(1)
				ss1.Add(4)
				ss2 := tt.ssc(nil)
				ss2.Add(4)
				ss2.Add(9)
				ss2.Union(ss1)
				want := []uint16{1, 4, 9}
				if diff := cmp.Diff(ss2.Elements(), want, cmpopts.SortSlices(func(a, b uint16) bool { return a < b })); diff != "" {
					t.Errorf("%s.Union().Elements() had diff %s", tt.name, diff)
				}
			})

			t.Run("Iter_Empty", func(t *testing.T) {
				ss := tt.ssc(nil)
				got := make([]uint16, 0)
				it := NewSeparationSetIterator(ss)
				for sep, ok := it.Next(); ok; sep, ok = it.Next() {
					got = append(got, sep)
				}
				want := []uint16{}
				if diff := cmp.Diff(got, want); diff != "" {
					t.Errorf("%s.Iter() had diff: %s", tt.name, diff)
				}
			})

			t.Run("Iter_Nonempty", func(t *testing.T) {
				ss := tt.ssc(grid.Placements{grid.Point{0, 0}, grid.Point{0, 1}, grid.Point{0, 3}})
				got := make([]uint16, 0)
				it := NewSeparationSetIterator(ss)
				for sep, ok := it.Next(); ok; sep, ok = it.Next() {
					got = append(got, sep)
				}
				want := []uint16{1, 4, 9}
				if diff := cmp.Diff(got, want, cmpopts.SortSlices(func(a, b uint16) bool { return a < b })); diff != "" {
					t.Errorf("%s.Iter() had diff: %s", tt.name, diff)
				}
			})

			t.Run("IterForGrid_Nonempty", func(t *testing.T) {
				ss := tt.ssc(grid.Placements{grid.Point{0, 0}, grid.Point{2, 2}, grid.Point{3, 3}})
				got := make([]uint16, 0)
				it := NewSeparationSetIteratorForGrid(ss, grid.Grid{3})
				for sep, ok := it.Next(); ok; sep, ok = it.Next() {
					got = append(got, sep)
				}
				want := []uint16{2, 8} // 18 is an invalid separation on a size 3 grid
				if diff := cmp.Diff(got, want, cmpopts.SortSlices(func(a, b uint16) bool { return a < b })); diff != "" {
					t.Errorf("%s.Iter() had diff: %s", tt.name, diff)
				}
			})
		})
	}
}

func Benchmark_BitArraySeparationSet(b *testing.B) {
	ss := NewBitArraySeparationSet(nil)
	for i := 0; i < b.N; i++ {
		ss.Clear()
		for sep := uint16(0); sep <= grid.MaxSeparation; sep++ {
			ss.Add(sep)
		}
		iter := NewSeparationSetIterator(ss)
		for sep, ok := iter.Next(); ok; sep, ok = iter.Next() {
			_ = sep
		}
	}
}

func Benchmark_BitArraySeparationSet_Iteration(b *testing.B) {
	// [A0 A1 A3 A7 B10 C6 F0 J9 L1 N3 N13]: 11 stones with unique separations on a 14x14 grid
	ss := NewBitArraySeparationSet(grid.Placements{
		grid.Point{0, 0},
		grid.Point{0, 1},
		grid.Point{0, 3},
		grid.Point{0, 7},
		grid.Point{1, 10},
		grid.Point{2, 6},
		grid.Point{5, 0},
		grid.Point{9, 9},
		grid.Point{11, 1},
		grid.Point{13, 3},
		grid.Point{13, 13},
	})
	for i := 0; i < b.N; i++ {
		iter := NewSeparationSetIterator(ss)
		for sep, ok := iter.Next(); ok; sep, ok = iter.Next() {
			_ = sep
		}
	}
}

func Test_bitSeparationSet_Clone_mapSeparationSet(t *testing.T) {
	sep1 := uint16(4)
	sep2 := uint16(6)
	ss1 := NewMapSeparationSet(nil)
	ss1.Add(sep1)
	ss2 := NewBitArraySeparationSet(nil)
	ss2.Add(sep2)
	ss2.Clone(ss1)
	if diff := cmp.Diff(ss1.Elements(), ss2.Elements()); diff != "" {
		t.Errorf("bitSeparationset.Clone(mapSeparationSet).Elements() had diff %s", diff)
	}
}

func Test_bitSeparationSet_Iteration(t *testing.T) {
	var got []uint16
	ss := NewBitArraySeparationSet(nil)
	ss.Add(5)  // Middle of a uint8
	ss.Add(15) // End of a uint8 before a 0 uint8
	ss.Add(56) // Beginning of a uint8 after a 0 uint8
	ss.Add(63)
	ss.Add(64) // Beginning of a uint8 after the end of previous one
	it := NewSeparationSetIterator(ss)
	for sep, ok := it.Next(); ok; sep, ok = it.Next() {
		got = append(got, sep)
	}
	want := []uint16{5, 15, 56, 63, 64}
	if diff := cmp.Diff(got, want, cmpopts.SortSlices(func(a, b uint16) bool { return a < b })); diff != "" {
		t.Errorf("Iter() had diff: %s", diff)
	}
}

func Test_bitSeparationSet_AdvanceCount(t *testing.T) {
	var got []uint16
	ss := NewBitArraySeparationSet(nil)
	ss.Add(5)
	it := SeparationSetIterator{SeparationSet: ss, maxSep: 20}
	// This loop mirrors the implementation of Next(), but doesn't stop if a separation is in the set
	for it.sep <= it.maxSep {
		got = append(got, it.sep)
		it.advance()
	}
	got = append(got, it.sep)

	want := []uint16{0, 1, 2, 3, 4, 5, 6, 7 /* Skip 2nd byte */, 16 /* Skip 3rd byte*/, 24 /* Above maxSep */}
	if diff := cmp.Diff(got, want); diff != "" {
		t.Errorf("advance() visited elements had diff (-got, +want): %s", diff)
	}
}

func Test_PointSet(t *testing.T) {
	// Arbitrary grid point values.
	point1 := grid.Point{Row: 1, Col: 2}
	point2 := grid.Point{Row: 3, Col: 4}
	point3 := grid.Point{Row: 5, Col: 6}

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

			t.Run("Union_Elements", func(t *testing.T) {
				// Add two different points to each set, then make the second set a clone of the first
				ps1 := tt.psc(grid.Placements{point1, point2})
				ps2 := tt.psc(grid.Placements{point1, point3})
				ps2.Union(ps1)
				want := grid.Placements{point1, point2, point3}
				if diff := cmp.Diff(ps2.Elements(), want, cmpopts.SortSlices(grid.LessThan)); diff != "" {
					t.Errorf("%s.Union().Elements() had diff %s", tt.name, diff)
				}
			})

			t.Run("Clear_Elements", func(t *testing.T) {
				ps := tt.psc(grid.Placements{point1, point2})
				ps.Clear()
				if got := len(ps.Elements()); got != 0 {
					t.Errorf("len(%s.Clear().Elements())=%d, want 0", tt.name, got)
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

func Test_bitArrayPointSet_Union_mapPointSet(t *testing.T) {
	// Arbitrary grid point values.
	point1 := grid.Point{Row: 1, Col: 2}
	point2 := grid.Point{Row: 3, Col: 4}
	ps1 := NewMapPointSet(nil)
	ps1.Add(point1)
	ps2 := NewBitArrayPointSet(nil)
	ps2.Add(point2)
	ps2.Union(ps1)
	if diff := cmp.Diff(ps2.Elements(), grid.Placements{point1, point2}); diff != "" {
		t.Errorf("bitArrayPointSet.Clone(mapPointSet).Elements() had diff %s", diff)
	}
}

func Test_bitArrayPointSet_MaxGridPoints(t *testing.T) {
	ps := NewBitArrayPointSet(nil)
	for row := uint8(0); row < grid.MaxGridSize; row++ {
		for col := uint8(0); col < grid.MaxGridSize; col++ {
			ps.Add(grid.Point{Row: row, Col: col})
		}
	}
	want := grid.MaxGridSize * grid.MaxGridSize
	if got := len(ps.Elements()); got != want {
		t.Errorf("Pointset has %d elements, want %d", got, want)
	}
}
