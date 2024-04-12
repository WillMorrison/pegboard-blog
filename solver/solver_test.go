package solver

import (
	"reflect"
	"testing"

	"github.com/WillMorrison/pegboard-blog/grid"
	"github.com/WillMorrison/pegboard-blog/placer"
)

func TestSingleOctantStartingPoints(t *testing.T) {
	tests := []struct {
		name string
		g    grid.Grid
		want []grid.Placements
	}{
		{"4x4",
			grid.Grid{Size: 4},
			[]grid.Placements{
				{grid.Point{0, 0}},
				{grid.Point{0, 1}},
				{grid.Point{1, 1}},
			},
		},
		{"5x5",
			grid.Grid{Size: 5},
			[]grid.Placements{
				{grid.Point{0, 0}},
				{grid.Point{0, 1}},
				{grid.Point{0, 2}},
				{grid.Point{1, 1}},
				{grid.Point{1, 2}},
				{grid.Point{2, 2}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SingleOctantStartingPoints(tt.g); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("SingleOctantStartingPoints() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSolver_Solve(t *testing.T) {

	tests := []struct {
		name   string
		solver Solver
	}{
		{"SingleThreadedSolver",
			SingleThreadedSolver{SingleOctantStartingPoints, placer.OrderedNoAllocStonePlacerProvider{}},
		},
		{"AsyncSolver",
			AsyncSolver{SingleOctantStartingPoints, placer.OrderedNoAllocStonePlacerProvider{}},
		},
		{"AsyncSplittingSolver",
			AsyncSplittingSolver{SingleOctantStartingPoints, placer.OrderedNoAllocStonePlacerProvider{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			t.Run("HasSolution", func(t *testing.T) {
				g := grid.Grid{Size: 7}
				got, err := tt.solver.Solve(g)
				if err != nil {
					t.Fatalf("%+v.Solve() error = %v", tt.solver, err)
					return
				}
				if err := grid.CheckValidSolution(g, got); err != nil {
					t.Errorf("%+v.Solve() = %v, want valid solution", tt.solver, got)
				}
			})

			t.Run("NoSolution", func(t *testing.T) {
				if testing.Short() {
					t.Skip("skipping test in short mode.")
				}
				g := grid.Grid{Size: 8}
				_, err := tt.solver.Solve(g)
				if err == nil {
					t.Errorf("%+v.Solve() error = %v: want err", tt.solver, err)
				}
			})
		})
	}
}
