package main

import (
	"fmt"

	"github.com/fatih/color"
)

// Color schemes for output:
//   - dirColor   — bold blue for directory names
//   - sizeColor  — green for sizes
//   - errColor   — red for permission error messages
//   - cleanColor — yellow for the "← can be cleaned" marker
var (
	dirColor   = color.New(color.FgBlue, color.Bold)
	sizeColor  = color.New(color.FgGreen)
	errColor   = color.New(color.FgRed)
	cleanColor = color.New(color.FgYellow)
)

// formatSize formats a size in bytes into a human-readable string.
// Automatically selects the unit: GB, MB, KB, or B.
// Fractional values are displayed with two decimal places.
func formatSize(b int64) string {
	const (
		KB = 1024
		MB = 1024 * KB
		GB = 1024 * MB
	)
	switch {
	case b >= GB:
		return fmt.Sprintf("%.2f GB", float64(b)/float64(GB))
	case b >= MB:
		return fmt.Sprintf("%.2f MB", float64(b)/float64(MB))
	case b >= KB:
		return fmt.Sprintf("%.2f KB", float64(b)/float64(KB))
	default:
		return fmt.Sprintf("%d B", b)
	}
}

// printTree recursively prints a directory tree node with tree(1)-style indentation.
//
// Parameters:
//   - node   — current node to print
//   - prefix — accumulated indent for the current level (│ separators or spaces)
//   - isLast — whether the node is the last among its siblings
//     (determines the connector: └── for the last, ├── for others)
//
// On a permission error (node.Err != nil) prints the directory name with [permission denied]
// and does not descend further.
func printTree(node *DirNode, prefix string, isLast bool) {
	connector := "├── "
	childPrefix := prefix + "│   "
	if isLast {
		connector = "└── "
		childPrefix = prefix + "    "
	}

	if node.Err != nil {
		fmt.Printf("%s%s%s  %s\n",
			prefix, connector,
			dirColor.Sprint(node.Name),
			errColor.Sprint("[permission denied]"),
		)
		return
	}

	marker := ""
	if cleanable[node.Path] {
		marker = "  " + cleanColor.Sprint("← can be cleaned")
	}

	fmt.Printf("%s%s%s  %s%s\n",
		prefix, connector,
		dirColor.Sprint(node.Name),
		sizeColor.Sprint(formatSize(node.Size)),
		marker,
	)

	for i, child := range node.Children {
		printTree(child, childPrefix, i == len(node.Children)-1)
	}
}
