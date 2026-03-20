# findspace

A CLI utility for visualizing a directory tree with sizes. Shows 3 levels of nesting, sizes in GB/MB/KB, sorted from largest to smallest. Marks directories that can be safely cleaned on Ubuntu.

## Installation on Ubuntu

```bash
go build -o findspace .
sudo mv findspace /usr/local/bin/
```

Now `findspace` is available from any directory.

## Usage

```
findspace -path <dir> -min <MB> [-clean 1]
```

| Flag | Default | Description |
|------|---------|-------------|
| `-path` | `.` | Directory to scan |
| `-min` | `0` | Hide directories smaller than N MB |
| `-clean` | `0` | `1` — prompt to clean cleanable directories after output |

```bash
findspace                               # current directory
findspace -path /home/user              # specific path
findspace -path /home/user -min 500     # show only >= 500 MB
findspace -path /home/user -clean 1     # show tree and offer cleanup
findspace -h                            # help
```

## Example output

```
/home/user  42.10 GB
├── .cache  18.39 GB  ← can be cleaned
│   ├── JetBrains  6.12 GB  ← can be cleaned
│   └── pip  6.61 GB  ← can be cleaned
├── projects  11.10 GB
│   └── backend  8.23 GB
└── .local  3.90 GB
    └── share
        └── Trash  1.20 GB  ← can be cleaned
```

With the `-clean 1` flag, a list and confirmation prompt are shown after the tree:

```
  /home/user/.cache  18.39 GB
  /home/user/.local/share/Trash  1.20 GB

Total space to free: 19.59 GB
Clean? [y/N]: y
Cleaned: 19.59 GB
```
