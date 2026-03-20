package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

// runClean prints the list of cleanable directories, asks for confirmation,
// and on approval deletes their contents via os.RemoveAll.
// After completion it prints the total amount of space freed.
func runClean() {
	type entry struct {
		path string
		size int64
	}

	var entries []entry
	var totalSize int64

	for path := range cleanable {
		if _, err := os.Stat(path); err == nil {
			sz := calcSize(path)
			if sz > 0 {
				entries = append(entries, entry{path, sz})
				totalSize += sz
			}
		}
	}

	if len(entries) == 0 {
		fmt.Println("\nNothing to clean — cleanable directories are empty or inaccessible.")
		return
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].size > entries[j].size
	})

	fmt.Println()
	for _, e := range entries {
		fmt.Printf("  %s  %s\n",
			cleanColor.Sprint(e.path),
			sizeColor.Sprint(formatSize(e.size)),
		)
	}
	fmt.Printf("\nTotal space to free: %s\n", sizeColor.Sprint(formatSize(totalSize)))
	fmt.Print("Clean? [y/N]: ")

	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))

	if answer != "y" {
		fmt.Println("Cancelled.")
		return
	}

	var freed int64
	for _, e := range entries {
		if err := os.RemoveAll(e.path); err != nil {
			fmt.Fprintf(os.Stderr, "error removing %s: %v\n", e.path, err)
		} else {
			freed += e.size
		}
	}

	const GB = 1024 * 1024 * 1024
	fmt.Printf("Cleaned: %.2f GB\n", float64(freed)/float64(GB))
}
