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
	header := fmt.Sprintf("Ran helm diff for %d charts:", len(chartDiffs))
	comment.WriteString(header)
	comment.WriteString("\n")

	for index, d := range chartDiffs {
		comment.WriteString(fmt.Sprintf("### %d. %s:", index+1, d.chartPath))
		comment.WriteString("\n")
		if d.hasChanged {
			comment.WriteString("```diff \n")
			comment.WriteString(d.diff)
			comment.WriteString("``` \n")
		} else {
			comment.WriteString("No diff detected. \n")
		}
	}
	return comment.String()
}
