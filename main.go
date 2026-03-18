package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/fatih/color"
)

type DirNode struct {
	Path     string
	Name     string
	Size     int64
	Children []*DirNode
	Err      error
}

func calcSize(path string) int64 {
	var total int64
	_ = filepath.WalkDir(path, func(_ string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			info, err := d.Info()
			if err == nil {
				total += info.Size()
			}
		}
		return nil
	})
	return total
}

func buildTree(path string, depth int) *DirNode {
	name := filepath.Base(path)
	if path == "/" {
		name = "/"
	}
	node := &DirNode{Path: path, Name: name}

	entries, err := os.ReadDir(path)
	if err != nil {
		node.Err = err
		return node
	}

	for _, entry := range entries {
		childPath := filepath.Join(path, entry.Name())
		if entry.IsDir() {
			if depth > 0 {
				child := buildTree(childPath, depth-1)
				node.Size += child.Size
				node.Children = append(node.Children, child)
			} else {
				sz := calcSize(childPath)
				node.Size += sz
				node.Children = append(node.Children, &DirNode{
					Path: childPath,
					Name: entry.Name(),
					Size: sz,
				})
			}
		} else {
			info, err := entry.Info()
			if err == nil {
				node.Size += info.Size()
			}
		}
	}

	sort.Slice(node.Children, func(i, j int) bool {
		return node.Children[i].Size > node.Children[j].Size
	})

	return node
}

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

var (
	dirColor  = color.New(color.FgBlue, color.Bold)
	sizeColor = color.New(color.FgGreen)
	errColor  = color.New(color.FgRed)
)

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

	fmt.Printf("%s%s%s  %s\n",
		prefix, connector,
		dirColor.Sprint(node.Name),
		sizeColor.Sprint(formatSize(node.Size)),
	)

	for i, child := range node.Children {
		printTree(child, childPrefix, i == len(node.Children)-1)
	}
}

func main() {
	path := "."
	if len(os.Args) > 1 {
		path = os.Args[1]
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	root := buildTree(abs, 2) // root + 3 levels below = depth 2 for children

	fmt.Printf("%s  %s\n",
		dirColor.Sprint(abs),
		sizeColor.Sprint(formatSize(root.Size)),
	)

	for i, child := range root.Children {
		printTree(child, "", i == len(root.Children)-1)
	}
}
