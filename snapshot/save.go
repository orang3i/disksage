package snapshot

import (
	"encoding/gob"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func Save(root string, dirSizes map[string]int64, files []FileEntry, snapshotPath string) error {

	if err := os.MkdirAll(snapshotPath, 0755); err != nil {
		return err
	}

	ts := time.Now().UTC()
	name := ts.Format("2006-01-02T15-04-05.000000000Z")
	name = strings.Replace(name, ".", "-", 1)
	filename := name + ".gob"

	tmpPath := filepath.Join(snapshotPath, filename+".tmp")
	finalPath := filepath.Join(snapshotPath, filename)

	file, err := os.Create(tmpPath)
	if err != nil {
		return err
	}

	snap := Snapshot{
		Header: SnapshotHeader{
			Version:   1,
			Timestamp: ts.Unix(),
			Root:      root,
		},
		DirSizes: dirSizes,
		Files:    files,
	}

	enc := gob.NewEncoder(file)
	if err := enc.Encode(&snap); err != nil {
		file.Close()
		_ = os.Remove(tmpPath)
		return err
	}

	if err := file.Sync(); err != nil {
		file.Close()
		_ = os.Remove(tmpPath)
		return err
	}

	if err := file.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}

	if err := os.Rename(tmpPath, finalPath); err != nil {
		_ = os.Remove(tmpPath)
		return err
	}

	return nil
}
