package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	"github.com/fatih/color"
)

// DirNode представляет узел дерева директорий.
// Size содержит суммарный размер всех файлов внутри директории рекурсивно,
// включая файлы в поддиректориях любой глубины.
// Err заполняется, если при чтении директории возникла ошибка (например, нет прав).
type DirNode struct {
	Path     string
	Name     string
	Size     int64
	Children []*DirNode
	Err      error
}

// calcSize возвращает суммарный размер всех файлов внутри директории path,
// обходя всё дерево рекурсивно через filepath.WalkDir.
// Ошибки доступа к отдельным файлам игнорируются — подсчёт продолжается.
// Используется для директорий, находящихся на максимальной глубине отображения,
// чтобы их размер был посчитан полностью, но дочерние узлы не создавались.
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

// buildTree рекурсивно строит дерево директорий начиная с path глубиной depth уровней.
//
// Логика работы:
//   - Читает содержимое директории path через os.ReadDir.
//   - Для каждого файла добавляет его размер к Size текущего узла.
//   - Для каждой поддиректории:
//   - Если depth > 0 — рекурсивно вызывает buildTree с depth-1, создавая дочерний узел.
//   - Если depth == 0 — считает размер через calcSize (без создания дочерних узлов),
//     чтобы не превышать заданную глубину отображения.
//
// Дочерние узлы сортируются по убыванию размера — самые тяжёлые папки отображаются первыми.
// При ошибке чтения директории узел возвращается с заполненным полем Err.
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

// Цветовые схемы для вывода:
//   - dirColor  — синий жирный для имён директорий
//   - sizeColor — зелёный для размеров
//   - errColor  — красный для сообщений об ошибках доступа
var (
	dirColor  = color.New(color.FgBlue, color.Bold)
	sizeColor = color.New(color.FgGreen)
	errColor  = color.New(color.FgRed)
)

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

	fmt.Printf("%s%s%s  %s\n",
		prefix, connector,
		dirColor.Sprint(node.Name),
		sizeColor.Sprint(formatSize(node.Size)),
	)

	for i, child := range node.Children {
		printTree(child, childPrefix, i == len(node.Children)-1)
	}
}

// main — точка входа. Принимает опциональный аргумент — путь к директории.
// Если аргумент не передан, использует текущую директорию.
// Строит дерево глубиной 3 уровня (корень + 3 уровня дочерних директорий)
// и выводит его в stdout с цветовым форматированием.
func main() {
	path := "."
	if len(os.Args) > 1 {
		path = os.Args[1]
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
