package snapshot

import (
	"encoding/gob"
	"os"
)

func Load(path string) (*Snapshot, error) {
	file, err := os.Open(path)

	if err != nil {
		return nil, err
	}
	defer file.Close()

	var snap Snapshot
	dec := gob.NewDecoder(file)

	if err := dec.Decode(&snap); err != nil {
		return nil, err
	}

	return &snap, nil
}
