package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

// main is the entry point.
//
// Flags:
//
//	-path string  directory to scan (default ".")
//	-min  int     hide directories smaller than N MB (default 0 — show all)
//
// Builds a 3-level directory tree and prints it to stdout with color formatting.
func main() {
	initCleanable()

	pathFlag := flag.String("path", ".", "directory to scan")
	minFlag := flag.Int64("min", 0, "hide directories smaller than N MB")
	cleanFlag := flag.Int("clean", 0, "1 — prompt to clean cleanable directories after output")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: findspace [flags]\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExample:\n")
		fmt.Fprintf(os.Stderr, "  findspace -path /home/user -min 500\n")
	}

	flag.Parse()

	if *cleanFlag != 0 && *cleanFlag != 1 {
		fmt.Fprintf(os.Stderr, "error: -clean must be 0 or 1\n")
		os.Exit(1)
	}
	if *minFlag < 0 {
		fmt.Fprintf(os.Stderr, "error: -min must be >= 0\n")
		os.Exit(1)
	}
	minSize = *minFlag * 1024 * 1024

	abs, err := filepath.Abs(*pathFlag)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// depth=2: root (level 0) → children (1) → grandchildren (2) → great-grandchildren sized via calcSize
	root := buildTree(abs, 2)

	fmt.Printf("%s  %s\n",
		dirColor.Sprint(abs),
		sizeColor.Sprint(formatSize(root.Size)),
	)

	for i, child := range root.Children {
		printTree(child, "", i == len(root.Children)-1)
	}

	if *cleanFlag == 1 {
		runClean()
	}
}
