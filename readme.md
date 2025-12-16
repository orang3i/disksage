```md
# disksage

`disksage` is a lightweight CLI tool to snapshot filesystem sizes and diff disk usage over time.

It helps answer questions like:
- Which directory grew?
- Which files changed size?
- What changed since the last scan?

Snapshots are stored as compact binary files and can be compared later.

---

## Configuration

### Where the config file lives

On first run, `disksage` automatically creates a configuration file at:

```

~/.config/disksage/config.json

````

This follows the XDG Base Directory specification on Linux.

You do not need to create this file manually.  
If it does not exist, `disksage` creates it with default values.

---

### Default configuration

A newly created config file looks like this:

```json
{
  "excluded_paths": [
    "/proc",
    "/sys",
    "/dev",
    "/run"
  ],
  "snapshotDir": "/home/user/.local/share/disksage/snapshots",
  "indexFiles": false
}
````

Explanation:

* **excluded_paths**
  Paths that are always skipped during scans. These are system directories that should never be scanned.

* **snapshotDir**
  Directory where snapshot files (`.gob`) are stored.

* **indexFiles**
  If `false`, only directory sizes are tracked.
  If `true`, individual file sizes are also stored and diffed.

---

### Snapshot storage location

By default, snapshots are stored in:

```
~/.local/share/disksage/snapshots
```

Each snapshot is stored as a binary file named using a timestamp, for example:

```
2025-12-16T20-20-14-472918332Z.gob
```

The snapshot directory is automatically excluded from scans, so snapshots never include themselves.

---

### Configuration precedence

Settings are resolved in the following order:

```
Command-line flags
↓
config.json
↓
built-in defaults
```

This means:

* Command-line flags override config values
* Config values override defaults
* The config file is never modified by CLI flags

---

## Usage

`disksage` is a single binary with subcommands:

```
disksage <command> [options]
```

Available commands:

* `scan` – create a snapshot
* `list` – list existing snapshots
* `diff` – compare snapshots

---

## Scan

Create a snapshot of directory sizes (and optionally file sizes).

### Basic scan

```
disksage scan
```

Scans the current directory and stores a snapshot in the default snapshot directory.

---

### Scan a specific path

```
disksage scan --path /home/user/projects
```

---

### Include individual file sizes

```
disksage scan --indexFiles
```

This allows diffs to show file-level changes.

---

### Override snapshot output directory

```
disksage scan --out /tmp/disksage-snaps
```

This stores snapshots in `/tmp/disksage-snaps` for this run only.

---

## List snapshots

List available snapshots and the root path they were taken from.

### Default snapshot directory

```
disksage list
```

Example output:

```
2025-12-16T20-20-14-472918332Z.gob  /home/user/projects
2025-12-16T20-25-01-998112332Z.gob  /home/user/projects
```

---

### List snapshots from a custom directory

```
disksage list --dir /tmp/disksage-snaps
```

---

## Diff snapshots

Compare snapshots and show what changed.

### Diff specific snapshots

```
disksage diff snapshot1.gob snapshot2.gob
```

---

### Diff the latest two snapshots

```
disksage diff --latest
```

---

### Diff using a custom snapshot directory

```
disksage diff --latest --dir /tmp/disksage-snaps
```

---

### Example diff output

```
=== DIRECTORY DIFF ===
+ /home/user/projects 1.3 MB

=== FILE DIFF ===
+ /home/user/projects/log.txt 120 KB
~ /home/user/projects/data.bin 5.0 MB → 7.0 MB
```

Explanation:

* A directory increased in size
* A new file was added
* A file grew in size

---

## Error handling behavior

* Permission errors are skipped (similar to `du` and `find`)
* System directories are excluded by default
* Scans do not fail due to unreadable files
* Errors are reported cleanly without panics

---

## Typical workflow

```
disksage scan --path /home/user
# wait some time
disksage scan --path /home/user
disksage diff --latest
```

This shows what changed between the two scans.

---

## Summary

* Config file: `~/.config/disksage/config.json`
* Snapshot storage: `~/.local/share/disksage/snapshots`
* CLI flags override config without modifying it
* Designed to behave like standard Unix tools

```
```
