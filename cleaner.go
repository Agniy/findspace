package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

// runClean выводит список cleanable-директорий, запрашивает подтверждение
// и при согласии удаляет их содержимое через os.RemoveAll.
// После завершения выводит суммарный объём освобождённого места.
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
		fmt.Println("\nНечего чистить — cleanable-директории пусты или недоступны.")
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
	fmt.Printf("\nИтого можно освободить: %s\n", sizeColor.Sprint(formatSize(totalSize)))
	fmt.Print("Очистить? [y/N]: ")

	reader := bufio.NewReader(os.Stdin)
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(strings.ToLower(answer))

	if answer != "y" && answer != "д" && answer != "да" {
		fmt.Println("Отменено.")
		return
	}

	var freed int64
	for _, e := range entries {
		if err := os.RemoveAll(e.path); err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка при удалении %s: %v\n", e.path, err)
		} else {
			freed += e.size
		}
	}

	const GB = 1024 * 1024 * 1024
	fmt.Printf("Очищено: %.2f GB\n", float64(freed)/float64(GB))
}
