package github

import (
	"fmt"
	"strings"

	"github.com/google/go-github/v74/github"
)

// CalculateDiffPosition converts a line number to a diff position for GitHub API
func CalculateDiffPosition(file *github.CommitFile, lineNumber int) int {
	patch := file.GetPatch()
	if patch == "" {
		return -1
	}

	lines := strings.Split(patch, "\n")
	position := 0
	newLineNumber := 0

	for _, line := range lines {
		if strings.HasPrefix(line, "@@") {
			// Parse hunk header: @@ -old_start,old_count +new_start,new_count @@
			newStart := parseHunkHeader(line)
			if newStart > 0 {
				newLineNumber = newStart - 1 // Adjust for 0-based indexing
			}
			position++
			continue
		}

		if strings.HasPrefix(line, " ") || strings.HasPrefix(line, "+") {
			// Context line or addition
			newLineNumber++
			position++

			if newLineNumber == lineNumber {
				return position
			}
		} else if strings.HasPrefix(line, "-") {
			// Deletion - increment position but not line number
			position++
		}
	}

	return -1 // Line not found in diff
}

// parseHunkHeader extracts the new file starting line number from a hunk header
func parseHunkHeader(hunkHeader string) int {
	// Format: @@ -old_start,old_count +new_start,new_count @@
	parts := strings.Split(hunkHeader, " ")
	if len(parts) < 3 {
		return 0
	}

	newPart := parts[2] // +new_start,new_count
	if !strings.HasPrefix(newPart, "+") {
		return 0
	}

	var newStart int
	fmt.Sscanf(newPart, "+%d", &newStart)
	return newStart
}

// GetChangedLines returns the line numbers that were changed in the file
func GetChangedLines(file *github.CommitFile) []int {
	patch := file.GetPatch()
	if patch == "" {
		return nil
	}

	var changedLines []int
	lines := strings.Split(patch, "\n")
	newLineNumber := 0

	for _, line := range lines {
		if strings.HasPrefix(line, "@@") {
			// Parse hunk header to get starting line number
			newStart := parseHunkHeader(line)
			if newStart > 0 {
				newLineNumber = newStart - 1
			}
			continue
		}

		if strings.HasPrefix(line, " ") {
			// Context line
			newLineNumber++
		} else if strings.HasPrefix(line, "+") {
			// Addition
			newLineNumber++
			changedLines = append(changedLines, newLineNumber)
		}
		// Deletions don't increment newLineNumber
	}

	return changedLines
}

// IsLineInDiff checks if a specific line number is part of the diff
func IsLineInDiff(file *github.CommitFile, lineNumber int) bool {
	changedLines := GetChangedLines(file)
	for _, line := range changedLines {
		if line == lineNumber {
			return true
		}
	}
	return false
}
