package cmd

import (
	"bytes"
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
	fileTree     *tview.TreeView
	fileTreeNode *tview.TreeNode
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

func (rp *RepoPage) Init() {
	rp.readmeView = tview.NewTextView()
	rp.readmeView.SetBorder(true)
	rp.readmeView.SetTitle("README.md")
	rp.readmeView.SetTitleAlign(tview.AlignLeft)
	rp.readmeView.SetText("Loading...")

	rp.fileTree = tview.NewTreeView()
	rp.fileTree.SetBorder(true)
	rp.fileTree.SetTitle("Files")
	rp.fileTree.SetTitleAlign(tview.AlignLeft)

	rp.Flex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(rp.fileTree, 0, 1, false).
		AddItem(rp.readmeView, 0, 1, false)

	rp.Flex.SetInputCapture(rp.onInputCapture)
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

		rp.fileTreeNode = tview.NewTreeNode("root")
		rp.fileTree.SetRoot(rp.fileTreeNode)
		rp.fileTree.SetCurrentNode(rp.fileTreeNode)

		// fetch the main readme.md of this repo
		for _, content := range rp.repo.contents {
			rp.fileTreeNode.AddChild(tview.NewTreeNode(content.Name))
			if strings.Index(strings.ToLower(content.Name), "readme.md") != -1 {
				rp.fetchFile(user, repo, rp.repo.Default_branch, content.Name)
			}
		}

		rp.readmeView.SetText(rp.fileContents)
	}()
}

func (rp *RepoPage) fetchRepo(user string, repo string) error {
	// get general repo properties
	response, err := FetchJson(fmt.Sprintf("https://api.github.com/repos/%s/%s", user, repo), &rp.repo)
	if err != nil {
		return err
	}
	if response.StatusCode == http.StatusNotFound {
		return errors.New("Repo doesn't exist. Press any button to return to the search page.")
	}

	// get repo contents metadata
	_, err = FetchJson(fmt.Sprintf("https://api.github.com/repos/%s/%s/contents", user, repo), &rp.repo.contents)
	if err != nil {
		return err
	}

	return nil
}

func (rp *RepoPage) fetchFile(user string, repo string, branch string, file string) error {
	response, err := Fetch(fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/refs/heads/%s/%s", user, repo, branch, file))
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)
	respBytes := buf.String()

	rp.fileContents = string(respBytes)

	return nil
}
