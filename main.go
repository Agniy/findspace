package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"sync"

	"github.com/fatih/color"
)

// minSize — минимальный размер директории в байтах для отображения в дереве.
// Директории меньше этого порога не попадают в Children, но их размер
// всё равно учитывается в суммарном размере родительского узла.
// Устанавливается из второго аргумента CLI (в MB); по умолчанию 0 (показывать всё).
var minSize int64

// sem — семафор для ограничения числа одновременно работающих горутин.
// Без ограничения при обходе широкого дерева можно исчерпать лимит файловых дескрипторов ОС.
// Размер буфера: CPU*8, т.к. операции I/O-bound и горутины большую часть времени ждут диск.
var sem = make(chan struct{}, runtime.NumCPU()*8)

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
// Используется для директорий на максимальной глубине отображения:
// их размер нужен полный, но дочерние узлы создавать не нужно.
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
// Параллелизм: каждая поддиректория обрабатывается в отдельной горутине.
// Семафор sem используется ТОЛЬКО на время самого I/O-вызова (os.ReadDir / calcSize),
// а не на всё время жизни горутины. Это критично: держать семафор во время wg.Wait()
// или рекурсивного buildTree приводит к дедлоку — дочерние горутины не могут получить
// слот, пока родительская горутина его удерживает в ожидании этих же дочерних.
//
// Логика по depth:
//   - depth > 0 — рекурсивно строим поддерево через buildTree(child, depth-1).
//   - depth == 0 — считаем размер через calcSize (полностью, но без дочерних узлов).
//
// Файлы (не директории) суммируются последовательно — их stat уже есть в entries.
// Дочерние узлы сортируются по убыванию размера.
// При ошибке чтения директории узел возвращается с заполненным полем Err.
func buildTree(path string, depth int) *DirNode {
	name := filepath.Base(path)
	if path == "/" {
		name = "/"
	}
	node := &DirNode{Path: path, Name: name}

	// Семафор занимаем только на время os.ReadDir, сразу освобождаем.
	sem <- struct{}{}
	entries, err := os.ReadDir(path)
	<-sem

	if err != nil {
		node.Err = err
		return node
	}

	// Размер файлов в текущей директории суммируем последовательно —
	// stat-вызовы быстрые, и эти данные уже есть в entries.
	var fileSize int64
	var dirs []os.DirEntry
	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry)
		} else {
			if info, err := entry.Info(); err == nil {
				fileSize += info.Size()
			}
		}
	}

	// Каждую поддиректорию обрабатываем в отдельной горутине.
	// Семафор НЕ держим во время wg.Wait() — иначе дедлок при рекурсии.
	var (
		mu       sync.Mutex
		wg       sync.WaitGroup
		children []*DirNode
		dirSize  int64
	)

	for _, entry := range dirs {
		childPath := filepath.Join(path, entry.Name())
		entryName := entry.Name()

		wg.Add(1)
		go func(cp, eName string) {
			defer wg.Done()

			var child *DirNode
			if depth > 0 {
				// Рекурсия управляет семафором самостоятельно внутри.
				child = buildTree(cp, depth-1)
			} else {
				// calcSize — чистый I/O без рекурсии горутин, семафор безопасен.
				sem <- struct{}{}
				sz := calcSize(cp)
				<-sem
				child = &DirNode{Path: cp, Name: eName, Size: sz}
			}

			mu.Lock()
			dirSize += child.Size
			// Добавляем в дерево только директории не меньше minSize.
			// Размер мелких директорий всё равно учитывается в dirSize родителя.
			if child.Size >= minSize {
				children = append(children, child)
			}
			mu.Unlock()
		}(childPath, entryName)
	}

	wg.Wait()

	node.Children = children
	node.Size = fileSize + dirSize

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
