package main

import (
	"flag"
	"github.com/Retropaint/ghterm/cmd"
)

func main() {
	cmd.Init()

	repo := flag.String("repo", "none", "Instantly opens a repo. \nExample: --repo retropaint/ghterm")
	flag.Parse()
	
	if *repo != "none" {
		cmd.OpenRepo(*repo)
	}

	cmd.Layout.Run()
}
