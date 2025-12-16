package main

import (
	"encoding/json"
	"flag"
	"io/fs"
	"log"
	"orang3i/disksage/snapshot"
	"os"
	"path/filepath"
	"strings"
)

func runScan(args []string) {
	fs := flag.NewFlagSet("scan", flag.ExitOnError)
	path := fs.String("path", ".", "root path to scan")
	indexFiles := fs.Bool("indexFiles", false, "index individual files")
	outDir := fs.String("out", "", "snapshot output directory (overrides config)")

	fs.Parse(args)

	if err := scan(*path, *indexFiles, *outDir); err != nil {
		log.Fatal(err)
	}
}

func scan(rootPath string, indexFiles bool, outDir string) error {
	cfg, err := loadConfig()
	if err != nil {
		return err
	}

	if indexFiles {
		cfg.IndexFiles = true
	}

	if outDir != "" {
		abs, err := filepath.Abs(outDir)
		if err != nil {
			return err
		}
		cfg.SnapshotDir = abs
	}

	excluded := make([]string, 0, len(cfg.ExcludedPaths))
	for _, p := range cfg.ExcludedPaths {
		excluded = append(excluded, filepath.Clean(p))
	}

	dirSizes := make(map[string]int64)
	files := []snapshot.FileEntry{}

	root, err := filepath.Abs(rootPath)
	if err != nil {
		return err
	}
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}

		if d.IsDir() {
			for _, ex := range excluded {
				if path == ex || strings.HasPrefix(path, ex+string(os.PathSeparator)) {
					return filepath.SkipDir
				}
			}
		}

		if d.Type()&fs.ModeSymlink != 0 {
			return nil
		}

		if d.IsDir() {
			if _, ok := dirSizes[path]; !ok {
				dirSizes[path] = 0
			}
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}

		size := info.Size()

		if cfg.IndexFiles {
			files = append(files, snapshot.FileEntry{
				Path: path,
				Size: size,
			})
		}

		parent := filepath.Dir(path)

		for {
			if !isUnderRoot(parent, root) {
				break
			}

			dirSizes[parent] += size

			if parent == root {
				break
			}

			next := filepath.Dir(parent)

			if next == parent {
				break
			}

			parent = next
		}
		return nil
	})
	if err != nil {
		return err
	}

	err = snapshot.Save(root, dirSizes, files, cfg.SnapshotDir)
	if err != nil {
		return err
	}

	return nil
}

func defaultConfigPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "disksage", "config.json"), nil
}

func defaultConfig() (*snapshot.Config, error) {
	dataDir, err := userDataDirFallback()
	if err != nil {
		return nil, err
	}
	return &snapshot.Config{
		ExcludedPaths: []string{
			"/proc",
			"/sys",
			"/dev",
			"/run",
		},
		SnapshotDir: filepath.Join(dataDir, "disksage", "snapshots"),
		IndexFiles:  false,
	}, nil
}

func userDataDirFallback() (string, error) {
	if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
		return xdg, nil
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share"), nil
}

func loadConfig() (*snapshot.Config, error) {
	cfg, err := defaultConfig()
	if err != nil {
		return nil, err
	}

	path, err := defaultConfigPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		// Create config if it doesnt exist
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return nil, err
		}

		b, err := json.MarshalIndent(cfg, "", "  ")
		if err != nil {
			return nil, err
		}

		if err := os.WriteFile(path, b, 0644); err != nil {
			return nil, err
		}

		return cfg, nil
	}

	var fileCfg snapshot.Config
	if err := json.Unmarshal(data, &fileCfg); err != nil {
		return nil, err
	}

	if len(fileCfg.ExcludedPaths) > 0 {
		cfg.ExcludedPaths = fileCfg.ExcludedPaths
	}
	if fileCfg.SnapshotDir != "" {
		cfg.SnapshotDir = fileCfg.SnapshotDir
	}
	cfg.IndexFiles = fileCfg.IndexFiles

	return cfg, nil
}

func isUnderRoot(path, root string) bool {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}
	return rel == "." || !strings.HasPrefix(rel, "..")
}
