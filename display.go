package main

import (
	"fmt"

	"github.com/fatih/color"
)

// Цветовые схемы для вывода:
//   - dirColor   — синий жирный для имён директорий
//   - sizeColor  — зелёный для размеров
//   - errColor   — красный для сообщений об ошибках доступа
//   - cleanColor — жёлтый для маркера "← можно почистить"
var (
	dirColor   = color.New(color.FgBlue, color.Bold)
	sizeColor  = color.New(color.FgGreen)
	errColor   = color.New(color.FgRed)
	cleanColor = color.New(color.FgYellow)
)

// formatSize форматирует размер в байтах в человекочитаемую строку.
// Автоматически выбирает единицу измерения: GB, MB, KB или B.
// Дробные значения выводятся с двумя знаками после запятой.
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

// printTree рекурсивно выводит узел дерева директорий с отступами в стиле tree(1).
//
// Параметры:
//   - node   — текущий узел для вывода
//   - prefix — накопленный отступ для текущего уровня (строки-разделители │ или пробелы)
//   - isLast — является ли узел последним среди своих соседей
//     (влияет на выбор коннектора: └── для последнего, ├── для остальных)
//
// При ошибке доступа (node.Err != nil) выводит имя директории и пометку [permission denied],
// не продолжая обход вглубь.
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
		marker = "  " + cleanColor.Sprint("← можно почистить")
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
