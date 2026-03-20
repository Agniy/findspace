package main

import (
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"sync"
)

// minSize is the minimum directory size in bytes to show in the tree.
// Directories below this threshold are excluded from Children, but their size
// is still counted toward the parent node's total.
// Set from the CLI flag in MB; default 0 (show all).
var minSize int64

// sem is a semaphore that limits the number of concurrently running goroutines.
// Without a limit, scanning a wide tree can exhaust the OS file descriptor limit.
// Buffer size: CPU*8, because operations are I/O-bound and goroutines mostly wait on disk.
var sem = make(chan struct{}, runtime.NumCPU()*8)

// DirNode represents a node in the directory tree.
// Size holds the total size of all files inside the directory recursively,
// including files in subdirectories at any depth.
// Err is set if an error occurred while reading the directory (e.g. permission denied).
type DirNode struct {
	Path     string
	Name     string
	Size     int64
	Children []*DirNode
	Err      error
}

// calcSize returns the total size of all files inside path,
// walking the entire tree recursively via filepath.WalkDir.
// Access errors for individual files are ignored — counting continues.
// Used for directories at the maximum display depth:
// their full size is needed, but child nodes do not need to be created.
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

// buildTree recursively builds a directory tree starting at path up to depth levels.
//
// Concurrency: each subdirectory is processed in a separate goroutine.
// The semaphore sem is held ONLY during the I/O call itself (os.ReadDir / calcSize),
// not for the entire goroutine lifetime. This is critical: holding the semaphore during
// wg.Wait() or recursive buildTree causes a deadlock — child goroutines cannot acquire
// a slot while the parent goroutine holds one waiting for those same children.
//
// Depth logic:
//   - depth > 0 — recursively build the subtree via buildTree(child, depth-1).
//   - depth == 0 — compute size via calcSize (full walk, no child nodes created).
//
// Files (non-directories) are summed sequentially — their stat is already in entries.
// Child nodes are sorted by descending size.
// On a directory read error the node is returned with Err set.
func buildTree(path string, depth int) *DirNode {
	name := filepath.Base(path)
	if path == "/" {
		name = "/"
	}
	node := &DirNode{Path: path, Name: name}

	// Acquire the semaphore only for the duration of os.ReadDir, then release immediately.
	sem <- struct{}{}
	entries, err := os.ReadDir(path)
	<-sem

	if err != nil {
		node.Err = err
		return node
	}

	// Sum file sizes in the current directory sequentially —
	// stat calls are fast and the data is already present in entries.
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

	// Process each subdirectory in a separate goroutine.
	// Do NOT hold the semaphore during wg.Wait() — that would cause a deadlock on recursion.
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
				// Recursion manages the semaphore internally.
				child = buildTree(cp, depth-1)
			} else {
				// calcSize is pure I/O without goroutine recursion, semaphore is safe here.
				sem <- struct{}{}
				sz := calcSize(cp)
				<-sem
				child = &DirNode{Path: cp, Name: eName, Size: sz}
			}

			mu.Lock()
			dirSize += child.Size
			// Only add directories of at least minSize to the tree.
			// Small directories still contribute to the parent's dirSize.
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
