package main

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

func TestFindCharts(t *testing.T) {
	tests := []struct {
		name   string
		charts []string // directories to create a Chart.yaml in, relative to root
		dir    string
		want   []string
	}{
		{
			name:   "dir is a chart",
			charts: []string{"chart"},
			dir:    "chart",
			want:   []string{"chart"},
		},
		{
			name:   "charts side by side",
			charts: []string{"charts/api", "charts/web"},
			dir:    "charts",
			want:   []string{"charts/api", "charts/web"},
		},
		{
			name:   "nested deeper",
			charts: []string{"charts/team-a/api", "charts/web"},
			dir:    "charts",
			want:   []string{"charts/team-a/api", "charts/web"},
		},
		{
			name:   "subcharts are skipped",
			charts: []string{"charts/api", "charts/api/charts/redis"},
			dir:    "charts",
			want:   []string{"charts/api"},
		},
		{
			name:   "outside dir is ignored",
			charts: []string{"charts/api", "other/web"},
			dir:    "charts",
			want:   []string{"charts/api"},
		},
		{
			name:   "no charts",
			charts: nil,
			dir:    ".",
			want:   nil,
		},
		{
			name:   "dir missing",
			charts: []string{"charts/api"},
			dir:    "deleted",
			want:   nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			root := t.TempDir()
			for _, c := range tt.charts {
				if err := os.MkdirAll(filepath.Join(root, c), 0o755); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(root, c, "Chart.yaml"), []byte("name: "+filepath.Base(c)+"\n"), 0o644); err != nil {
					t.Fatal(err)
				}
			}

			got, err := findCharts(root, tt.dir)
			if err != nil {
				t.Fatalf("findCharts() error = %v", err)
			}
			if !slices.Equal(got, tt.want) {
				t.Errorf("findCharts() = %q, want %q", got, tt.want)
			}
		})
	}
}
