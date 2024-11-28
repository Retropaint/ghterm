package main

import (
	"flag"
	"github.com/Retropaint/ghterm/internal"
)

func main() {
	internal.Init()

	isCommits := flag.Bool("commits", false, "See commits for repository defined by --repo")
	repo := flag.String("repo", "none", "Points to a repo for other flags, or opens it's main page if none are provided. \nExample: --repo retropaint/ghterm")
	flag.Parse()

	if *repo != "none" {
		if *isCommits {
			internal.OpenCommits(*repo)
		} else {
			internal.OpenRepo(*repo)
		}
	}

	internal.Layout.Run()
}
