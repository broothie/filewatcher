package main

import (
	"fmt"
	"os"

	"github.com/alecthomas/kingpin"
	"github.com/broothie/filewatcher"
)

var (
	cmd  = kingpin.Arg("command", "command to run on file change").Required().String()
	glob = kingpin.Flag("glob", "files to watch").Short('g').Default("*").String()
	root = kingpin.Flag("root", "directory root to watch from").Short('r').Default(".").String()
)

func init() {
	kingpin.Parse()
}

func main() {
	fw, err := filewatcher.New(*cmd, *glob, *root)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	fw.Start()
}
