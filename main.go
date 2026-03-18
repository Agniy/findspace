package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
)

// main — точка входа.
//
// Флаги:
//
//	-path string  директория для обхода (по умолчанию ".")
//	-min  int     скрывать директории меньше N MB (по умолчанию 0 — показывать всё)
//
// Строит дерево глубиной 3 уровня и выводит его в stdout с цветовым форматированием.
func main() {
	initCleanable()

	pathFlag := flag.String("path", ".", "директория для обхода")
	minFlag := flag.Int64("min", 0, "скрыть директории меньше N MB")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Использование: findspace [флаги]\n\n")
		fmt.Fprintf(os.Stderr, "Флаги:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nПример:\n")
		fmt.Fprintf(os.Stderr, "  findspace -path /home/user -min 500\n")
	}

	flag.Parse()

	if *minFlag < 0 {
		fmt.Fprintf(os.Stderr, "error: -min должен быть >= 0\n")
		os.Exit(1)
	}
	minSize = *minFlag * 1024 * 1024

	abs, err := filepath.Abs(*pathFlag)
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
