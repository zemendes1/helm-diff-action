package main

import "testing"

func TestFormatComment(t *testing.T) {
	tests := []struct {
		name       string
		chartDiffs []chartDiff
		want       string
	}{{name: "two charts unchanged",
		chartDiffs: []chartDiff{{
			chartPath:  "charts/web",
			diff:       "",
			hasChanged: false}, {
			chartPath:  "charts/api",
			diff:       "",
			hasChanged: false}},
		want: "Ran helm diff for 2 charts:\n### 1. charts/web:\nNo diff detected.\n### 2. charts/api:\nNo diff detected."}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatComment(tt.chartDiffs)
			if got != tt.want {
				t.Errorf("want: %q, got: %q", tt.want, got)
			}
		})
	}
}
