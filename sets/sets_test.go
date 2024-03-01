package sets

import (
	"testing"

	"github.com/WillMorrison/pegboard-blog/grid"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

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
