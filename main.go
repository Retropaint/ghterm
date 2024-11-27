package main

import (
	"flag"
	"github.com/Retropaint/ghterm/internal"
)

func main() {
	internal.Init()

	repo := flag.String("repo", "none", "Instantly opens a repo. \nExample: --repo retropaint/ghterm")
	commits := flag.String("commits", "none", "Instantly opens commits for repoo. \nExample: --commits retropaint/ghterm")
	flag.Parse()

	if *repo != "none" {
		internal.OpenRepo(*repo)
	}
	if *commits != "none" {
		internal.OpenCommits(*commits)
	}

	internal.Layout.Run()
}
