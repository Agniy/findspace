package main

import (
	"os"
	"path/filepath"
)

// cleanable is the set of absolute paths that are safe to delete on Ubuntu
// to free disk space. Populated by initCleanable().
var cleanable map[string]bool

// initCleanable builds the cleanable set: expands ~ to the real home directory
// and adds both user caches and system temporary directories.
//
// Sources: Ubuntu Community Help Wiki (RecoverLostDiskSpace),
// APT, Snap, Flatpak, JetBrains, npm, pip documentation.
func initCleanable() {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	j := filepath.Join

	paths := []string{
		// User caches — entirely safe to delete;
		// applications recreate them on next launch.
		j(home, ".cache"),
		j(home, ".cache", "JetBrains"),     // IDE caches for all versions
		j(home, ".cache", "pip"),           // pip HTTP cache
		j(home, ".cache", "google-chrome"), // browser cache
		j(home, ".cache", "mozilla"),       // Firefox cache
		j(home, ".cache", "sublime-text"),  // Sublime Text cache

		// Package manager caches — safe to delete, rebuilt on demand.
		j(home, ".npm"),
		j(home, ".pip"),
		j(home, ".gradle"),
		j(home, ".composer"),
		j(home, ".m2", "repository"), // Maven local repository

		// User trash.
		j(home, ".local", "share", "Trash"),

		// Flatpak — update cache.
		j(home, ".local", "share", "flatpak", "cache"),

		// System temporary directories — cleaned by the OS, but can be cleared manually.
		"/tmp",
		"/var/tmp",

		// APT — downloaded .deb packages. Remove via `sudo apt clean`.
		"/var/cache/apt",
		"/var/cache/apt/archives",

		// Snap — old package revisions (the active revision is not touched).
		"/var/lib/snapd/cache",
	}

	cleanable = make(map[string]bool, len(paths))
	for _, p := range paths {
		cleanable[p] = true
	}
}
