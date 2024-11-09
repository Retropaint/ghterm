package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type RepoPage struct {
	*tview.Flex
	readmeView   *tview.TextView
	fileContents string
	repo         Repo
	repo404      bool
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

	rp.Flex.SetInputCapture(rp.onInputCapture)

	if name != "none" {
		rp.GetRepo(name)
	}
}

func (rp *RepoPage) onInputCapture(event *tcell.EventKey) *tcell.EventKey {
	if rp.repo404 {
		Layout.Pages.SwitchToPage("search")
	}

	return event
}

func (rp *RepoPage) GetRepo(name string) {
	user := strings.Split(name, "/")[0]
	repo := strings.Split(name, "/")[1]
	go func() {
		defer Layout.App.Draw()

		err := rp.fetchRepo(user, repo)
		if err != nil {
			rp.readmeView.SetText(fmt.Sprint(err))
			rp.repo404 = true
			return
		}

		// fetch the main readme.md of this repo
		for _, content := range rp.repo.contents {
			if strings.Index(strings.ToLower(content.Name), "readme.md") != -1 {
				rp.fetchFile(user, repo, rp.repo.Default_branch, content.Name)
			}
		}

		rp.readmeView.SetText(rp.fileContents)
	}()
}

func (rp *RepoPage) fetchRepo(user string, repo string) error {
	// get general repo properties
	response := Fetch("https://api.github.com/repos/" + user + "/" + repo)
	err := json.NewDecoder(response.Body).Decode(&rp.repo)
	if err != nil {
		return err
	}

	if response.StatusCode == http.StatusNotFound {
		return errors.New("Repo doesn't exist. Press any button to return to the search page.")
	}

	// get repo contents metadata
	response = Fetch("https://api.github.com/repos/" + user + "/" + repo + "/contents")
	err = json.NewDecoder(response.Body).Decode(&rp.repo.contents)
	if err != nil {
		return err
	}

	return nil
}

func (rp *RepoPage) fetchFile(user string, repo string, branch string, file string) {
	response := Fetch("https://raw.githubusercontent.com/" + user + "/" + repo + "/refs/heads/" + branch + "/" + file)

	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)
	respBytes := buf.String()

	rp.fileContents = string(respBytes)
}
