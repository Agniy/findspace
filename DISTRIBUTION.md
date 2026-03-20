# Distribution Options for Ubuntu

## Option 1: GitHub Releases (simplest)

Build the binary and attach it to a GitHub release. Users download and install manually.

```bash
# Build for linux/amd64
GOOS=linux GOARCH=amd64 go build -o findspace .

# Create archive
tar -czf findspace-linux-amd64.tar.gz findspace

# Upload to GitHub via gh CLI
gh release create v1.0.0 findspace-linux-amd64.tar.gz \
  --title "v1.0.0" \
  --notes "Initial release"
```

Installation for users:

```bash
curl -L https://github.com/Agniy/findspace/releases/latest/download/findspace-linux-amd64.tar.gz | tar xz
sudo mv findspace /usr/local/bin/
```

---

## Option 2: .deb Package (native Ubuntu)

**Package structure:**

```
findspace_1.0.0_amd64/
├── DEBIAN/
│   └── control
└── usr/
    └── local/
        └── bin/
            └── findspace
```

```bash
# Create structure
mkdir -p findspace_1.0.0_amd64/DEBIAN
mkdir -p findspace_1.0.0_amd64/usr/local/bin

# Build binary
GOOS=linux GOARCH=amd64 go build -o findspace_1.0.0_amd64/usr/local/bin/findspace .
chmod +x findspace_1.0.0_amd64/usr/local/bin/findspace
```

`findspace_1.0.0_amd64/DEBIAN/control`:

```
Package: findspace
Version: 1.0.0
Section: utils
Priority: optional
Architecture: amd64
Maintainer: Your Name <you@example.com>
Description: Find large directories on disk
 Utility for finding large directories on disk with color output.
```

```bash
# Build .deb
dpkg-deb --build findspace_1.0.0_amd64

# Result: findspace_1.0.0_amd64.deb
```

Installation for users:

```bash
wget https://github.com/Agniy/findspace/releases/latest/download/findspace_1.0.0_amd64.deb
sudo dpkg -i findspace_1.0.0_amd64.deb
```

---

## Option 3: Snap (recommended for broad coverage)

Snap works on all Ubuntu/Debian-based distributions and is published to the Snap Store.

**`snap/snapcraft.yaml`:**

```yaml
name: findspace
base: core22
version: '1.0.0'
summary: Find large directories on disk
description: |
  Utility for finding large directories with color output and
  markers for known cache directories that can be safely cleaned.

grade: stable
confinement: classic

parts:
  findspace:
    plugin: go
    source: .

apps:
  findspace:
    command: bin/findspace
```

```bash
# Install snapcraft
sudo snap install snapcraft --classic

# Build snap
snapcraft

# Result: findspace_1.0.0_amd64.snap
```

**Publishing to Snap Store:**

```bash
snapcraft login
snapcraft register findspace
snapcraft upload findspace_1.0.0_amd64.snap --release=stable
```

Installation for users:

```bash
sudo snap install findspace --classic
```

---

## Option 4: PPA (Launchpad) — via `apt install`

The most "official" approach — users add a repository and install via apt. More complex to set up: requires GPG signing and an account on [launchpad.net](https://launchpad.net).

Full documentation: [help.launchpad.net/Packaging/PPA](https://help.launchpad.net/Packaging/PPA)

---

## Recommended path

| Stage | Action |
|-------|--------|
| Now | GitHub Releases + .deb file |
| Later | Snap Store — `sudo snap install findspace` |
| If it grows | PPA for `apt install` |

For a utility of this scale, **GitHub Releases with a .deb** is the optimal start. Adding Snap is straightforward and gives broad coverage.
