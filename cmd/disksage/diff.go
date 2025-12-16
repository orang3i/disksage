package main

import (
	"flag"
	"fmt"
	"orang3i/disksage/snapshot"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type DirDiff struct {
	Path  string
	Delta int64
}

type FileDiff struct {
	Path    string
	OldSize int64
	NewSize int64
}

func runDiff(args []string) {
	fs := flag.NewFlagSet("diff", flag.ExitOnError)
	latest := fs.Bool("latest", false, "diff latest two snapshots")
	fs.Parse(args)

	var oldPath, newPath string

	if *latest {
		cfg, err := loadConfig()
		if err != nil {
			panic(err)
		}
		oldPath, newPath, err = latestTwoSnapshots(cfg.SnapshotDir)
		if err != nil {
			panic(err)
		}
	} else {
		rest := fs.Args()
		if len(rest) != 2 {
			fmt.Println("Usage: disksage diff <old_snapshot> <new_snapshot>")
			fmt.Println("   or: disksage diff --latest")
			return
		}
		oldPath, newPath = rest[0], rest[1]
	}

	oldSnap, err := snapshot.Load(oldPath)
	if err != nil {
		panic(err)
	}

	newSnap, err := snapshot.Load(newPath)
	if err != nil {
		panic(err)
	}

	root := oldSnap.Header.Root

	fmt.Println("=== DIRECTORY DIFF ===")
	dirDiffs := diffDirs(oldSnap.DirSizes, newSnap.DirSizes)
	printDirDiffs(root, dirDiffs)

	if len(oldSnap.Files) > 0 || len(newSnap.Files) > 0 {
		fmt.Println("\n=== FILE DIFF ===")
		fileDiffs := diffFiles(oldSnap.Files, newSnap.Files)
		printFileDiffs(root, fileDiffs)
	}
}

func diffDirs(oldDirs, newDirs map[string]int64) []DirDiff {
	seen := make(map[string]bool)

	diffs := []DirDiff{}

	for path, newSize := range newDirs {
		oldSize := oldDirs[path]
		if newSize != oldSize {
			diffs = append(diffs, DirDiff{
				Path:  path,
				Delta: newSize - oldSize,
			})
		}
		seen[path] = true
	}

	for path, oldSize := range oldDirs {
		if seen[path] {
			continue
		}
		diffs = append(diffs, DirDiff{
			Path:  path,
			Delta: -oldSize,
		})
	}

	sort.Slice(diffs, func(i, j int) bool {
		return abs(diffs[i].Delta) > abs(diffs[j].Delta)
	})

	return diffs
}

func diffFiles(oldFiles, newFiles []snapshot.FileEntry) []FileDiff {
	oldMap := make(map[string]int64)
	newMap := make(map[string]int64)

	for _, f := range oldFiles {
		oldMap[f.Path] = f.Size
	}
	for _, f := range newFiles {
		newMap[f.Path] = f.Size
	}

	seen := make(map[string]bool)
	diffs := []FileDiff{}

	for path, newSize := range newMap {
		oldSize, ok := oldMap[path]
		if !ok {
			diffs = append(diffs, FileDiff{
				Path:    path,
				OldSize: 0,
				NewSize: newSize,
			})
		} else if oldSize != newSize {
			diffs = append(diffs, FileDiff{
				Path:    path,
				OldSize: oldSize,
				NewSize: newSize,
			})
		}
		seen[path] = true
	}

	for path, oldSize := range oldMap {
		if seen[path] {
			continue
		}
		diffs = append(diffs, FileDiff{
			Path:    path,
			OldSize: oldSize,
			NewSize: 0,
		})
	}

	sort.Slice(diffs, func(i, j int) bool {
		return abs(diffs[i].NewSize-diffs[i].OldSize) >
			abs(diffs[j].NewSize-diffs[j].OldSize)
	})

	return diffs
}

func printDirDiffs(root string, diffs []DirDiff) {
	for _, d := range diffs {
		sign := "+"
		if d.Delta < 0 {
			sign = "-"
		}
		fmt.Printf("%s %s %s\n", sign, absPath(root, d.Path), humanSize(abs(d.Delta)))

	}
}

func printFileDiffs(root string, diffs []FileDiff) {
	for _, d := range diffs {
		path := absPath(root, d.Path)

		if d.OldSize == 0 {
			fmt.Printf("+ %s %s\n", path, humanSize(d.NewSize))
		} else if d.NewSize == 0 {
			fmt.Printf("- %s %s\n", path, humanSize(d.OldSize))
		} else {
			fmt.Printf(
				"~ %s %s → %s\n",
				path,
				humanSize(d.OldSize),
				humanSize(d.NewSize),
			)
		}
	}
}

func abs(x int64) int64 {
	if x < 0 {
		return -x
	}
	return x
}

func absPath(root, p string) string {
	if p == "." || p == "" {
		return root
	}
	return filepath.Join(root, p)
}

func latestTwoSnapshots(dir string) (string, string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", "", err
	}

	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.HasSuffix(e.Name(), ".gob") {
			names = append(names, e.Name())
		}
	}

	if len(names) < 2 {
		return "", "", fmt.Errorf("need at least 2 snapshots in %s", dir)
	}

	sort.Strings(names)
	oldName := names[len(names)-2]
	newName := names[len(names)-1]

	return filepath.Join(dir, oldName), filepath.Join(dir, newName), nil
}

func humanSize(n int64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := int64(unit), 0
	for n >= div*unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(n)/float64(div), "KMGTPE"[exp])
}
