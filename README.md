# findspace

Консольная утилита для визуализации дерева папок с размерами. Показывает 3 уровня вложенности, размеры в GB/MB/KB, сортирует от большего к меньшему.

## Установка на Ubuntu

```bash
go build -o findspace .
sudo mv findspace /usr/local/bin/
```

Теперь `findspace` доступен из любой папки.

## Использование

```
findspace [path] [min_mb]
```

| Аргумент | По умолчанию | Описание |
|----------|-------------|----------|
| `path`   | `.`         | Директория для обхода |
| `min_mb` | `0`         | Скрывать директории меньше N MB |

```bash
findspace                  # текущая папка
findspace /home/user       # конкретный путь
findspace /home/user 500   # показать только >= 500 MB
```

## Пример вывода

```
/home/user/projects  12.45 GB
├── backend          8.23 GB
│   ├── vendor       6.10 GB
│   ├── cmd          1.05 GB
│   └── pkg          1.08 GB
└── frontend         3.90 GB
    └── node_modules 3.75 GB
```
