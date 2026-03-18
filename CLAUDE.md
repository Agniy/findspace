# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
go build -o findspace .     # сборка бинарника
go run . [флаги]            # запуск без сборки
go mod tidy                 # обновить зависимости
```

## Usage

```
findspace -path <dir> -min <MB>
```

| Флаг | По умолчанию | Описание |
|------|-------------|----------|
| `-path` | `.` | Директория для обхода |
| `-min` | `0` | Скрывать директории меньше N MB |

## Установка (Ubuntu)

```bash
go build -o findspace .
sudo mv findspace /usr/local/bin/
```

## Architecture

Четыре файла в пакете `main`:

| Файл | Содержимое |
|------|-----------|
| `main.go` | Парсинг флагов (`-path`, `-min`), вызов `initCleanable` + `buildTree` + `printTree` |
| `tree.go` | `DirNode`, `buildTree`, `calcSize`, переменные `sem` и `minSize` |
| `display.go` | `printTree`, `formatSize`, цветовые переменные |
| `cleanable.go` | `cleanable map[string]bool`, `initCleanable` |

Поток выполнения:

1. `main` устанавливает `minSize` (байты), вызывает `buildTree(abs, 2)`
2. `buildTree` рекурсивно строит дерево `DirNode` глубиной 3 уровня (depth 2 → 1 → 0):
   - На каждом уровне запускает горутину на каждую поддиректорию
   - На `depth == 0` вызывает `calcSize` (полный обход без создания узлов)
   - Узлы меньше `minSize` не добавляются в `Children`, но их размер суммируется в родителе
   - Дочерние узлы сортируются по убыванию размера
3. `printTree` рекурсивно выводит дерево; если путь узла есть в `cleanable`, добавляет жёлтый маркер `← можно почистить`

### Важное про семафор

`sem` (`NumCPU*8`) захватывается **только на время самого I/O-вызова** (`os.ReadDir` / `calcSize`) и сразу освобождается. Держать семафор во время `wg.Wait()` или рекурсивного `buildTree` вызывает дедлок — родитель ждёт детей, дети не могут получить слот.

### cleanable

`initCleanable` строит `map[string]bool` из известных Ubuntu-путей (кеши пользователя, корзина, APT, Snap, Flatpak, пакетные менеджеры). Список целиком в `cleanable.go` — туда же добавлять новые пути.
