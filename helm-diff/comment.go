package main

import (
	"fmt"
	"strings"
)

type chartDiff struct {
	chartPath  string
	diff       string
	hasChanged bool
}

func formatComment(chartDiffs []chartDiff) string {
	var comment strings.Builder

	if len(chartDiffs) == 1 {
		fmt.Fprintf(&comment, "Ran helm diff for %d chart:", len(chartDiffs))
	} else {
		fmt.Fprintf(&comment, "Ran helm diff for %d charts:", len(chartDiffs))
	}

	for index, d := range chartDiffs {
		fmt.Fprintf(&comment, "\n### %d. %s:", index+1, d.chartPath)
		if d.hasChanged {
			comment.WriteString("```diff \n")
			comment.WriteString(d.diff)
			comment.WriteString("``` \n")
		} else {
			comment.WriteString("\nNo diff detected.")
		}
	}
	return comment.String()
}
