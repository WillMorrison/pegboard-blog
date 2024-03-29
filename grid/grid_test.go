package grid

import (
	"reflect"
	"slices"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestPoint_String(t *testing.T) {
	tests := []struct {
		p    Point
		want string
	}{
		{Point{0, 0}, "A0"},
		{Point{4, 2}, "E2"},
	}
	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.p.String(); got != tt.want {
				t.Errorf("Point.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsInBounds(t *testing.T) {
	type args struct {
		g Grid
		p Point
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{"top left", args{Grid{5}, Point{0, 0}}, true},
		{"bottom right", args{Grid{5}, Point{4, 4}}, true},
		{"outside right", args{Grid{5}, Point{4, 5}}, false},
		{"outside bottom", args{Grid{5}, Point{6, 3}}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsInBounds(tt.args.g, tt.args.p); got != tt.want {
				t.Errorf("IsInBounds() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSeparation(t *testing.T) {
	type args struct {
		p1 Point
		p2 Point
	}
	tests := []struct {
		name string
		args args
		want uint16
	}{
		{"p2 equals p1", args{Point{3, 2}, Point{3, 2}}, 0},
		{"p2 right of p1", args{Point{0, 0}, Point{0, 3}}, 9},
		{"p2 below p1", args{Point{0, 0}, Point{3, 0}}, 9},
		{"p2 left of p1", args{Point{0, 3}, Point{0, 0}}, 9},
		{"p2 above p1", args{Point{0, 3}, Point{0, 0}}, 9},
		{"p2 right and below p1", args{Point{0, 0}, Point{2, 3}}, 13},
		{"p2 left and above p1", args{Point{4, 5}, Point{2, 3}}, 8},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Separation(tt.args.p1, tt.args.p2); got != tt.want {
				t.Errorf("Separation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestCheckValidSolution(t *testing.T) {
	type args struct {
		g Grid
		p Placements
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{"valid 3x3",
			args{
				Grid{3},
				Placements{Point{0, 0}, Point{1, 1}, Point{1, 2}},
			},
			false},
		{"invalid 3x3 not enough stones",
			args{
				Grid{3},
				Placements{Point{0, 0}, Point{1, 1}},
			},
			true},
		{"invalid 3x3 out of bounds stone",
			args{
				Grid{3},
				Placements{Point{0, 0}, Point{1, 1}, Point{0, 4}},
			},
			true},
		{"invalid 2x2 colliding stones",
			args{
				Grid{2},
				Placements{Point{0, 0}, Point{0, 0}},
			},
			true},
		{"invalid 3x3 duplicate separations",
			args{
				Grid{3},
				Placements{Point{0, 0}, Point{1, 1}, Point{0, 2}},
			},
			true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := CheckValidSolution(tt.args.g, tt.args.p); tt.wantErr == (got == nil) {
				t.Errorf("CheckValidSolution() error = %v, want %v", got, tt.wantErr)
			}
		})
	}
}

func TestPlacements_Sort(t *testing.T) {
	tests := []struct {
		name string
		p    Placements
		want Placements
	}{
		{"Empty",
			Placements{},
			Placements{}},
		{"Descending",
			Placements{Point{1, 2}, Point{1, 1}, Point{0, 2}},
			Placements{Point{0, 2}, Point{1, 1}, Point{1, 2}}},
		{"Already sorted",
			Placements{Point{0, 0}, Point{1, 1}, Point{1, 1}},
			Placements{Point{0, 0}, Point{1, 1}, Point{1, 1}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := slices.Clone[Placements](tt.p)
			p.Sort()
			if !cmp.Equal(p, tt.want) {
				t.Errorf("%v.Sort() got %v want %v", tt.p, p, tt.want)
			}
		})
	}
}

func TestAdvanceStone(t *testing.T) {
	type args struct {
		g Grid
		p Point
	}
	tests := []struct {
		name string
		args args
		want Point
	}{
		{name: "along row", args: args{g: Grid{5}, p: Point{1, 2}}, want: Point{1, 3}},
		{name: "end of row", args: args{g: Grid{5}, p: Point{1, 4}}, want: Point{2, 0}},
		{name: "end of grid", args: args{g: Grid{5}, p: Point{4, 4}}, want: Point{5, 0}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := AdvanceStone(tt.args.g, tt.args.p); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AdvanceStone() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGrid_Iter(t *testing.T) {
	g := Grid{2}
	it := g.Iter()
	var got Placements
	for p, ok := it.Next(); ok; p, ok = it.Next() {
		got = append(got, p)
	}
	want := Placements{Point{0, 0}, Point{0, 1}, Point{1, 0}, Point{1, 1}}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Iter() produced %v, want %v", got, want)
	}
}
