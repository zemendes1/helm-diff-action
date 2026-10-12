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
		want: "Ran helm diff for 2 charts:\n### 1. charts/web:\nNo diff detected.\n### 2. charts/api:\nNo diff detected."},
		{name: "one chart changed one chart unchanged",
			chartDiffs: []chartDiff{{
				chartPath:  "charts/web",
				diff:       "",
				hasChanged: false}, {
				chartPath:  "charts/api",
				diff:       "--- main\n+++ HEAD\n@@ -5,5 +5,5 @@\n metadata:\n   name: release-name-config\n data:\n-  message: \"hello\"\n+  message: \"hello from the PR\"\n   env: \"prod\"\n",
				hasChanged: true}},
			want: "Ran helm diff for 2 charts:\n### 1. charts/web:\nNo diff detected.\n### 2. charts/api:\n```diff\n--- main\n+++ HEAD\n@@ -5,5 +5,5 @@\n metadata:\n   name: release-name-config\n data:\n-  message: \"hello\"\n+  message: \"hello from the PR\"\n   env: \"prod\"\n```"}}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatComment(tt.chartDiffs)
			if got != tt.want {
				t.Errorf("got:\n%s\n\ngot (quoted): %q\nwant (quoted): %q", got, got, tt.want)
			}
		})
	}
}
