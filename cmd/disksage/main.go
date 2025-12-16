package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "scan":
		runScan(os.Args[2:])
	case "diff":
		runDiff(os.Args[2:])
	case "list":
		runList(os.Args[2:])
	case "-h", "--help":
		usage()
	default:
		fmt.Printf("Unknown command: %s\n\n", os.Args[1])
		usage()
		os.Exit(1)
	}
}

func usage() {
	fmt.Println(`disksage - disk usage snapshot tool

Usage:
  disksage scan --path <dir> [--out <snapshot-dir>]
  disksage diff <old_snapshot> <new_snapshot>
  disksage list [--dir <snapshot-dir>]

Commands:
  scan    Scan filesystem and create snapshot
  diff    Compare two snapshots
  list	  List snapshots

Run 'disksage <command> --help' for details.`)
}
