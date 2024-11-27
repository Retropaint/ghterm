package internal

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type CommitsPage struct {
	*tview.Flex
	commitsList *tview.TextView
	commits     []Commit
	commitsIdx  int
}

type Commit struct {
	Author CommitAuthor
	Commit CommitBody
}

type CommitAuthor struct {
	Login string
}

type CommitBody struct {
	Message string
}

func (c *CommitsPage) Init() {
	c.commitsList = tview.NewTextView()
	c.commitsList.SetDynamicColors(true)
	c.commitsList.SetText("Loading...")
	c.commitsList.SetRegions(true)
	c.commitsList.SetInputCapture(c.onInputCapture)

	c.Flex = tview.NewFlex().
		AddItem(c.commitsList, 0, 1, true)
	c.Flex.SetBorder(true)
	c.Flex.SetTitle("Commits")
	c.Flex.SetTitleAlign(tview.AlignLeft)
}

func (c *CommitsPage) GetCommits(repo string) {
	go func() {
		c.fetchCommits(repo)
		c.populateList()
		Layout.App.Draw()
	}()
}

func (c *CommitsPage) populateList() {
	c.commitsList.SetText("")
	for i, co := range c.commits {
		fmt.Fprint(c.commitsList, fmt.Sprintf("[\"%d\"]", i))
		fmt.Fprintln(c.commitsList, co.Commit.Message)
		fmt.Fprintln(c.commitsList, "[aqua]"+co.Author.Login+"[-]")
		fmt.Fprintln(c.commitsList)
		fmt.Fprint(c.commitsList, "[\"\"]")
	}
	c.commitsList.Highlight("0")
}

func (c *CommitsPage) onInputCapture(event *tcell.EventKey) *tcell.EventKey {
	switch event.Name() {
	case "Rune[j]":
		c.commitsIdx = min(c.commitsIdx+1, len(c.commits)-1)
	case "Rune[k]":
		c.commitsIdx = max(c.commitsIdx-1, 0)
	case "Enter":
	}
	c.commitsList.Highlight(strconv.Itoa(c.commitsIdx))
	c.commitsList.ScrollToHighlight()
	return event
}

func (c *CommitsPage) fetchCommits(repo string) error {
	user := strings.Split(repo, "/")[0]
	repoName := strings.Split(repo, "/")[1]
	_, err := FetchJson(fmt.Sprintf("https://api.github.com/repos/%s/%s/commits", user, repoName), &c.commits)
	if err != nil {
		return err
	}
	return nil
}
