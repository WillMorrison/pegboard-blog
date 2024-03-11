package pruner

import (
	"reflect"
	"testing"

	"github.com/WillMorrison/pegboard-blog/grid"
	"github.com/WillMorrison/pegboard-blog/sets"
)

func Test_Pruner_PruneIsoceles(t *testing.T) {
	tests := []struct {
		name string
		grid grid.Grid
		p1   grid.Point
		p2   grid.Point
		want grid.Placements
	}{
		{
			name: "1,1 diagonal",
			grid: grid.Grid{5},
			p1:   grid.Point{0, 1},
			p2:   grid.Point{1, 0},
			want: grid.Placements{grid.Point{0, 0}, grid.Point{1, 1}, grid.Point{2, 2}, grid.Point{3, 3}, grid.Point{4, 4}},
		},
		{
			name: "horizontal with points",
			grid: grid.Grid{5},
			p1:   grid.Point{0, 0},
			p2:   grid.Point{2, 0},
			want: grid.Placements{grid.Point{1, 0}, grid.Point{1, 1}, grid.Point{1, 2}, grid.Point{1, 3}, grid.Point{1, 4}},
		},
		{
			name: "horizontal no points",
			grid: grid.Grid{5},
			p1:   grid.Point{0, 0},
			p2:   grid.Point{0, 1},
			want: grid.Placements{},
		},
		{
			name: "vertical with points",
			grid: grid.Grid{5},
			p1:   grid.Point{0, 0},
			p2:   grid.Point{0, 4},
			want: grid.Placements{grid.Point{0, 2}, grid.Point{1, 2}, grid.Point{2, 2}, grid.Point{3, 2}, grid.Point{4, 2}},
		},
		{
			name: "vertical no points",
			grid: grid.Grid{5},
			p1:   grid.Point{0, 0},
			p2:   grid.Point{1, 0},
			want: grid.Placements{},
		},
	}
	impls := []struct {
		name string
		new  func(grid.Grid) Pruner
	}{
		{name: "runtime", new: NewRuntimePruner},
		{name: "precomputed", new: NewPrecomputedPruner},
	}
	for _, impl := range impls {
		for _, tt := range tests {
			t.Run(impl.name+"/"+tt.name, func(t *testing.T) {
				p := impl.new(tt.grid)
				ps := sets.BitArrayPointSet{} // This implementation always returns ordered Elements()
				p.PruneIsoceles(&ps, tt.p1, tt.p2)
				if got := ps.Elements(); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("PruneIsoceles(%s, %s) = %v, want %v", tt.p1, tt.p2, got, tt.want)
				}
			})
		}
	}
}

func Test_Pruner_PruneCircles(t *testing.T) {
	tests := []struct {
		name string
		grid grid.Grid
		p1   grid.Point
		sep  uint16
		want grid.Placements
	}{
		{
			name: "possible separation, middle of grid",
			grid: grid.Grid{5},
			p1:   grid.Point{2, 2},
			sep:  1,
			want: grid.Placements{grid.Point{1, 2}, grid.Point{2, 1}, grid.Point{2, 3}, grid.Point{3, 2}},
		},
		{
			name: "possible separation, edge of grid",
			grid: grid.Grid{5},
			p1:   grid.Point{0, 0},
			sep:  1,
			want: grid.Placements{grid.Point{0, 1}, grid.Point{1, 0}},
		},
		{
			name: "impossible separation",
			grid: grid.Grid{5},
			p1:   grid.Point{2, 2},
			sep:  3,
			want: grid.Placements{},
		},
		{
			name: "pythagorean triple",
			grid: grid.Grid{6},
			p1:   grid.Point{0, 0},
			sep:  25, // could be 0+25 or 9+16
			want: grid.Placements{grid.Point{0, 5}, grid.Point{3, 4}, grid.Point{4, 3}, grid.Point{5, 0}},
		},
	}
	impls := []struct {
		name string
		new  func(grid.Grid) Pruner
	}{
		{name: "runtime", new: NewRuntimePruner},
		{name: "precomputed", new: NewPrecomputedPruner},
	}
	for _, impl := range impls {
		for _, tt := range tests {
			t.Run(impl.name+"/"+tt.name, func(t *testing.T) {
				p := impl.new(tt.grid)
				ps := sets.BitArrayPointSet{} // This implementation always returns ordered Elements()
				p.PruneCircles(&ps, tt.p1, tt.sep)
				if got := ps.Elements(); !reflect.DeepEqual(got, tt.want) {
					t.Errorf("PruneCircles(%s, %d) = %v, want %v", tt.p1, tt.sep, got, tt.want)
				}
			})
		}
	}
}

func Benchmark_PrecomputedPruner(b *testing.B) {
	g := grid.Grid{7}
	stones := grid.Placements{grid.Point{0, 0}, grid.Point{0, 2}, grid.Point{1, 2}, grid.Point{2, 6}, grid.Point{3, 0}, grid.Point{5, 5}, grid.Point{6, 6}}
	
	p := NewPrecomputedPruner(g)
	pruned := sets.BitArrayPointSet{}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pruned.Clear()
		for i, p1 := range stones {
			if pruned.Has(p1) {
				b.Fatalf("cannot place stone #%d at %s", i, p1)
			}
			for j := 0; j < i; j++ {
				p2 := stones[j]
				sep := grid.Separation(p1, p2)
				p.PruneIsoceles(&pruned, p1, p2)
				p.PruneCircles(&pruned, p1, sep)
			}
		}
	}
}
