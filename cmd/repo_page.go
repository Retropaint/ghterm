package cmd

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/rivo/tview"
)

type RepoPage struct {
	*tview.Flex
	readmeView   *tview.TextView
	fileContents string
	repo         Repo
}

type Repo struct {
	Name           string
	Full_name      string
	Description    string
	Default_branch string
	contents       []RepoContent
}

type RepoContent struct {
	Name string
	Path string
}

func (rp *RepoPage) Init(name string) {
	rp.readmeView = tview.NewTextView()
	rp.readmeView.SetBorder(true)
	rp.readmeView.SetTitle("README.md")
	rp.readmeView.SetTitleAlign(tview.AlignLeft)
	rp.readmeView.SetText("Loading...")

	rp.Flex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(rp.readmeView, 0, 1, false)

	if name != "none" {
		rp.GetRepo(name)
	}
}

func (rp *RepoPage) GetRepo(name string) {
	user := strings.Split(name, "/")[0]
	repo := strings.Split(name, "/")[1]
	go func() {
		rp.fetchRepo(user, repo)
		for _, content := range rp.repo.contents {
			if strings.Index(strings.ToLower(content.Name), "readme.md") != -1 {
				rp.fetchFile(user, repo, rp.repo.Default_branch, content.Name)
			}
		}
		rp.readmeView.SetText(rp.fileContents)
		Layout.App.Draw()
	}()
}

func (rp *RepoPage) fetchRepo(user string, repo string) {
	// get main repo fields
	response := Fetch("https://api.github.com/repos/" + user + "/" + repo)
	err := json.NewDecoder(response.Body).Decode(&rp.repo)
	if err != nil {
	}

	// get repo contents metadata
	response = Fetch("https://api.github.com/repos/" + user + "/" + repo + "/contents")
	err = json.NewDecoder(response.Body).Decode(&rp.repo.contents)
	if err != nil {
	}
}

func (rp *RepoPage) fetchFile(user string, repo string, branch string, file string) {
	response := Fetch("https://raw.githubusercontent.com/" + user + "/" + repo + "/refs/heads/" + branch + "/" + file)

	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)
	respBytes := buf.String()

	rp.fileContents = string(respBytes)
}
