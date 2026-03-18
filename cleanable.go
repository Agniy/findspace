package main

import (
	"os"
	"path/filepath"
)

// cleanable — множество абсолютных путей, которые безопасно удалять на Ubuntu
// для освобождения дискового пространства. Заполняется в initCleanable().
var cleanable map[string]bool

// initCleanable строит множество cleanable: раскрывает ~ в реальный домашний каталог
// и добавляет как пользовательские кеши, так и системные временные директории.
//
// Источники: Ubuntu Community Help Wiki (RecoverLostDiskSpace),
// документация APT, Snap, Flatpak, JetBrains, npm, pip.
func initCleanable() {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	j := filepath.Join

	paths := []string{
		// Кеши пользователя — целиком безопасны для удаления,
		// приложения пересоздают их при следующем запуске.
		j(home, ".cache"),
		j(home, ".cache", "JetBrains"),     // кеши IDE всех версий
		j(home, ".cache", "pip"),           // pip http-кеш
		j(home, ".cache", "google-chrome"), // кеш браузера
		j(home, ".cache", "mozilla"),       // кеш Firefox
		j(home, ".cache", "sublime-text"),  // кеш Sublime Text

		// Кеши пакетных менеджеров — безопасны, пересчитываются по требованию.
		j(home, ".npm"),
		j(home, ".pip"),
		j(home, ".gradle"),
		j(home, ".composer"),
		j(home, ".m2", "repository"), // Maven локальный репозиторий

		// Корзина пользователя.
		j(home, ".local", "share", "Trash"),

		// Flatpak — кеш обновлений.
		j(home, ".local", "share", "flatpak", "cache"),

		// Системные временные директории — очищаются ОС, но можно и вручную.
		"/tmp",
		"/var/tmp",

		// APT — скачанные .deb-пакеты. Удалять через `sudo apt clean`.
		"/var/cache/apt",
		"/var/cache/apt/archives",

		// Snap — старые ревизии пакетов (активная ревизия не трогается).
		"/var/lib/snapd/cache",
	}

	cleanable = make(map[string]bool, len(paths))
	for _, p := range paths {
		cleanable[p] = true
	}
}
