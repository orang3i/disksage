package snapshot

type Config struct {
	ExcludedPaths []string `json:"excluded_paths"`
	SnapshotDir   string   `json:"snapshotDir"`
	IndexFiles    bool     `json:"indexFiles"`
}

type SnapshotHeader struct {
	Version   uint16
	Timestamp int64
	Root      string
}

type Snapshot struct {
	Header   SnapshotHeader
	DirSizes map[string]int64
	Files    []FileEntry
}

type FileEntry struct {
	Path string
	Size int64
}

type ScanResult struct {
	DirSizes map[string]int64
	Files    []FileEntry
}
