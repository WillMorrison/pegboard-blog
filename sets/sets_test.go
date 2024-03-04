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
		})
	}
}

// Arbitrary grid point values.
var (
	point1 = grid.Point{Row: 1, Col: 2}
	point2 = grid.Point{Row: 3, Col: 4}
)

func Test_mapPointSet_Has(t *testing.T) {
	ps := NewMapPointSet(nil)
	if ps.Has(point1) {
		t.Errorf("%s.Has(%s)=true, want false", ps, point1)
	}
}

func Test_mapPointSet_AddHas(t *testing.T) {
	ps := NewMapPointSet(nil)
	ps.Add(point1)
	if !ps.Has(point1) {
		t.Errorf("%s.Has(%s)=false, want true", ps, point1)
	}
}

func Test_mapPointSet_AddHasOther(t *testing.T) {
	ps := NewMapPointSet(nil)
	ps.Add(point1)
	if ps.Has(point2) {
		t.Errorf("%s.Has(%s)=true, want false", ps, point2)
	}
}

func Test_mapPointSet_AddCopyHas(t *testing.T) {
	ps1 := NewMapPointSet(nil)
	ps1.Add(point1)
	ps2 := ps1.Copy()
	if !ps2.Has(point1) {
		t.Errorf("%s.Has(%s)=false, want true", ps2, point1)
	}
}

func Test_mapPointSet_AddCopyHasOther(t *testing.T) {
	ps1 := NewMapPointSet(nil)
	ps1.Add(point1)
	ps2 := ps1.Copy()
	if ps2.Has(point2) {
		t.Errorf("%s.Has(%s)=true, want false", ps2, point2)
	}
}

func Test_mapPointSet_CopyAddHas(t *testing.T) {
	ps1 := NewMapPointSet(nil)
	ps2 := ps1.Copy()
	ps2.Add(point1)
	if ps1.Has(point1) {
		t.Errorf("%s.Has(%s)=true, want false", ps1, point1)
	}
}

func Test_mapPointSet_Elements(t *testing.T) {
	tests := []struct {
		name string
		ps   PointSet
		want grid.Placements
	}{
		{
			name: "nil",
			ps:   NewMapPointSet(nil),
			want: grid.Placements{},
		},
		{
			name: "empty",
			ps:   NewMapPointSet(grid.Placements{}),
			want: grid.Placements{},
		},
		{
			name: "nonempty",
			ps:   NewMapPointSet(grid.Placements{point1, point2}),
			want: grid.Placements{point1, point2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.ps.Elements(); !cmp.Equal(got, tt.want, cmpopts.SortSlices(grid.LessThan)) {
				t.Errorf("mapPointSet.Elements() = %v, want %v", got, tt.want)
			}
		})
	}
}
