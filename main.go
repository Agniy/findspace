package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// main — точка входа.
//
// Аргументы:
//
//	findspace [path] [min_mb]
//	  path   — директория для обхода (по умолчанию ".")
//	  min_mb — минимальный размер поддиректории в MB для отображения (по умолчанию 0)
//
// Строит дерево глубиной 3 уровня и выводит его в stdout с цветовым форматированием.
func main() {
	initCleanable()

	path := "."
	if len(os.Args) > 1 {
		path = os.Args[1]
	}

	if len(os.Args) > 2 {
		mb, err := strconv.ParseInt(os.Args[2], 10, 64)
		if err != nil || mb < 0 {
			_, _ = fmt.Fprintf(os.Stderr, "error: min_mb must be a non-negative integer, got %q\n", os.Args[2])
			os.Exit(1)
		}
		minSize = mb * 1024 * 1024
	}

	abs, err := filepath.Abs(path)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	// depth=2: корень (уровень 0) → дочерние (1) → внуки (2) → правнуки считаются через calcSize
	root := buildTree(abs, 2)

	fmt.Printf("%s  %s\n",
		dirColor.Sprint(abs),
		sizeColor.Sprint(formatSize(root.Size)),
	)

	for i, child := range root.Children {
		printTree(child, "", i == len(root.Children)-1)
	}
}
