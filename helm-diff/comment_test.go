package main

import (
	"slices"
	"strings"
	"testing"
)

func TestFormatComment(t *testing.T) {
	tests := []struct {
		name       string
		chartDiffs []chartDiff
		want       []string
	}{{name: "two charts unchanged",
		chartDiffs: []chartDiff{{
			chartPath:  "charts/web",
			diff:       "",
			hasChanged: false}, {
			chartPath:  "charts/api",
			diff:       "",
			hasChanged: false}},
		want: []string{"Ran helm diff for 2 charts:\n### 1. charts/web:\nNo diff detected.\n### 2. charts/api:\nNo diff detected."}},
		{name: "one chart changed one chart unchanged",
			chartDiffs: []chartDiff{{
				chartPath:  "charts/web",
				diff:       "",
				hasChanged: false}, {
				chartPath:  "charts/api",
				diff:       "--- main\n+++ HEAD\n@@ -5,5 +5,5 @@\n metadata:\n   name: release-name-config\n data:\n-  message: \"hello\"\n+  message: \"hello from the PR\"\n   env: \"prod\"\n",
				hasChanged: true}},
			want: []string{"Ran helm diff for 2 charts:\n### 1. charts/web:\nNo diff detected.\n### 2. charts/api:\n```diff\n--- main\n+++ HEAD\n@@ -5,5 +5,5 @@\n metadata:\n   name: release-name-config\n data:\n-  message: \"hello\"\n+  message: \"hello from the PR\"\n   env: \"prod\"\n```"}},
		{name: "values files in headings",
			chartDiffs: []chartDiff{{
				chartPath:  "charts/api",
				values:     "values-prod.yaml",
				hasChanged: false}, {
				chartPath:  "charts/web",
				hasChanged: false}},
			want: []string{"Ran helm diff for 2 charts:\n### 1. charts/api / values-prod.yaml:\nNo diff detected.\n### 2. charts/web:\nNo diff detected."}},
		{name: "diff without trailing newline",
			chartDiffs: []chartDiff{{
				chartPath:  "charts/api",
				diff:       "--- main\n+++ HEAD\n@@ -1 +1 @@\n-  env: \"prod\"\n+  env: \"staging\"",
				hasChanged: true}},
			want: []string{"Ran helm diff for 1 chart:\n### 1. charts/api:\n```diff\n--- main\n+++ HEAD\n@@ -1 +1 @@\n-  env: \"prod\"\n+  env: \"staging\"\n```"}}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatComment(tt.chartDiffs)
			if !slices.Equal(got, tt.want) {
				t.Errorf("got:\n%s\n\ngot (quoted): %q\nwant (quoted): %q", got, got, tt.want)
			}
		})
	}
}

// bigDiff returns a diff of roughly size characters, made of whole lines.
func bigDiff(size int) string {
	line := "+  message: \"hello from the PR\"\n"
	return "--- main\n+++ HEAD\n" + strings.Repeat(line, size/len(line))
}

func TestFormatCommentSplitting(t *testing.T) {
	tests := []struct {
		name        string
		chartDiffs  []chartDiff
		minComments int
		maxComments int
	}{
		{name: "one 150,000-character diff",
			chartDiffs: []chartDiff{{
				chartPath:  "charts/api",
				diff:       bigDiff(150_000),
				hasChanged: true}},
			minComments: 3,
			maxComments: 4},
		{name: "three 40,000-character diffs",
			chartDiffs: []chartDiff{{
				chartPath:  "charts/web",
				diff:       bigDiff(40_000),
				hasChanged: true}, {
				chartPath:  "charts/api",
				diff:       bigDiff(40_000),
				hasChanged: true}, {
				chartPath:  "charts/worker",
				diff:       bigDiff(40_000),
				hasChanged: true}},
			minComments: 2,
			maxComments: 3},
		{name: "small diff fits in one comment",
			chartDiffs: []chartDiff{{
				chartPath:  "charts/web",
				diff:       "",
				hasChanged: false}, {
				chartPath:  "charts/api",
				diff:       bigDiff(1_000),
				hasChanged: true}},
			minComments: 1,
			maxComments: 1}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatComment(tt.chartDiffs)
			if len(got) < tt.minComments || len(got) > tt.maxComments {
				t.Errorf("got %d comments, want between %d and %d", len(got), tt.minComments, tt.maxComments)
			}
			for i, c := range got {
				if len(c) > maxCharacterSize {
					t.Errorf("comment %d is %d characters, want at most %d", i+1, len(c), maxCharacterSize)
				}
				if n := strings.Count(c, "```"); n%2 != 0 {
					t.Errorf("comment %d has %d code fences, want an even number:\n%s", i+1, n, c[max(0, len(c)-200):])
				}
			}
		})
	}
}

// Chart 1 leaves a little room at the end of the first comment; sweeping its
// size makes sure chart 2's heading never lands there with an empty diff block.
func TestFormatCommentNoEmptyDiffBlock(t *testing.T) {
	for size := 65_400; size <= 65_460; size++ {
		got := formatComment([]chartDiff{{
			chartPath:  "charts/web",
			diff:       strings.Repeat("+", size) + "\n",
			hasChanged: true}, {
			chartPath:  "charts/api",
			diff:       bigDiff(1_000),
			hasChanged: true}})
		for i, c := range got {
			if strings.Contains(c, "```diff\n```") {
				t.Fatalf("chart 1 diff of %d characters: comment %d has an empty diff block:\n%s", size, i+1, c[max(0, len(c)-100):])
			}
		}
	}
}
