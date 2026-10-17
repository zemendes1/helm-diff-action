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

const maxCharacterSize = 65536

func formatComment(chartDiffs []chartDiff) []string {
	var currentComment strings.Builder
	var comments []string

	if len(chartDiffs) == 1 {
		fmt.Fprintf(&currentComment, "Ran helm diff for %d chart:", len(chartDiffs))
	} else {
		fmt.Fprintf(&currentComment, "Ran helm diff for %d charts:", len(chartDiffs))
	}

	openDiff := ("\n```diff\n")
	closeDiff := ("```")
	newLine := "\n"
	noDiff := "\nNo diff detected."

	symbolsLenght := len(openDiff) + len(closeDiff) + len(newLine)

	for index, d := range chartDiffs {
		heading := fmt.Sprintf("\n### %d. %s:", index+1, d.chartPath)

		headingSize := len(heading) + len(noDiff)
		if d.hasChanged {
			firstLine, _, _ := strings.Cut(d.diff, newLine)
			headingSize = len(heading) + len(openDiff) + len(firstLine) + len(newLine) + symbolsLenght
		}
		if currentComment.Len()+headingSize > maxCharacterSize {
			comments = append(comments, currentComment.String())
			currentComment.Reset()
		}
		currentComment.WriteString(heading)

		if d.hasChanged {
			currentComment.WriteString(openDiff)
			for line := range strings.Lines(d.diff) {
				if currentComment.Len()+len(line)+symbolsLenght > maxCharacterSize {
					currentComment.WriteString(closeDiff)
					comments = append(comments, currentComment.String())
					currentComment.Reset()

					currentComment.WriteString(openDiff)
				}
				currentComment.WriteString(line)
			}
			if !strings.HasSuffix(d.diff, newLine) {
				currentComment.WriteString(newLine)
			}
			currentComment.WriteString(closeDiff)
		} else {
			currentComment.WriteString(noDiff)
		}
	}
	comments = append(comments, currentComment.String())
	return comments
}
