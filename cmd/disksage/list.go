package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"orang3i/disksage/snapshot"
)

func runList(args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	dir := fs.String("dir", "", "snapshot directory (overrides config)")
	fs.Parse(args)

	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, "config:", err)
		os.Exit(1)
	}

	snapshotDir := cfg.SnapshotDir
	if *dir != "" {
		abs, err := filepath.Abs(*dir)
		if err != nil {
			fmt.Fprintln(os.Stderr, "dir:", err)
			os.Exit(1)
		}
		snapshotDir = abs
	}

	entries, err := os.ReadDir(snapshotDir)
	if err != nil {
		fmt.Fprintln(os.Stderr, "readdir:", err)
		os.Exit(1)
	}

	type item struct {
		name string
		path string
	}

	items := make([]item, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if !strings.HasSuffix(e.Name(), ".gob") {
			continue
		}
		full := filepath.Join(snapshotDir, e.Name())
		items = append(items, item{name: e.Name(), path: full})
	}

	sort.Slice(items, func(i, j int) bool {
		return items[i].name < items[j].name
	})

	for _, it := range items {
		snap, err := snapshot.Load(it.path)
		if err != nil {
			fmt.Printf("%s  (unreadable: %v)\n", it.name, err)
			continue
		}
		fmt.Printf("%s  %s\n", it.name, snap.Header.Root)
	}
}
