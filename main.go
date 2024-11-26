package main

import (
	"flag"
	"github.com/Retropaint/ghterm/internal"
)

func main() {
	internal.Init()

	repo := flag.String("repo", "none", "Instantly opens a repo. \nExample: --repo retropaint/ghterm")
	flag.Parse()

	if *repo != "none" {
		internal.OpenRepo(*repo)
	}

	internal.Layout.Run()
}
