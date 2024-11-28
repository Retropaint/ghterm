package internal

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type CommitPage struct {
	*tview.Flex
	fileTree     *tview.TreeView
	fileTreeNode *tview.TreeNode
	patch        *tview.TextView
	commit       Commit
}

type Commit struct {
	Files []CommitFiles
}

type CommitFiles struct {
	Filename  string
	Patch     string
	Additions int
	Deletions int
}

func (c *CommitPage) Init() {
	c.fileTree = tview.NewTreeView()
	c.fileTreeNode = tview.NewTreeNode("root")
	c.fileTree.SetRoot(c.fileTreeNode)
	c.fileTree.SetCurrentNode(c.fileTreeNode)
	c.fileTree.SetBorder(true)

	c.patch = tview.NewTextView()
	c.patch.SetBorder(true)
	c.patch.SetDynamicColors(true)

	c.Flex = tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(c.fileTree, 0, 1, false).
		AddItem(c.patch, 0, 1, false)
	c.Flex.SetBorder(true)

	c.Flex.SetInputCapture(c.onInputCapture)
	c.fileTree.SetInputCapture(c.fileTreeOnInputCapture)
}

func (c *CommitPage) OpenCommit(repo string, sha string, name string) {
	c.patch.SetText("Loading...")
	go func() {
		err := c.fetchCommit(repo, sha)
		if err != nil {
			c.patch.SetText(err.Error())
			return
		}

		// add color codes to patches
		for i, f := range c.commit.Files {
			c.fileTreeNode.AddChild(tview.NewTreeNode(f.Filename))
			f := &c.commit.Files[i].Patch
			*f = strings.ReplaceAll(*f, "\n", "[-]\n")
			*f = strings.ReplaceAll(*f, "\n-", "[red]\n-")
			*f = strings.ReplaceAll(*f, "\n+", "[green]\n+")
			*f = strings.ReplaceAll(*f, "@@\n", "[-]@@\n")
			*f = strings.ReplaceAll(*f, "@@", "[purple]@@")
		}

		c.patch.SetText(c.commit.Files[0].Patch)
		Layout.App.Draw()
	}()
}

func (c *CommitPage) onInputCapture(event *tcell.EventKey) *tcell.EventKey {
	switch event.Name() {
	case "Ctrl+F":
		Layout.App.SetFocus(c.fileTree)
	case "Ctrl+L":
		Layout.App.SetFocus(c.patch)

	}

	return event
}

func (c *CommitPage) fileTreeOnInputCapture(event *tcell.EventKey) *tcell.EventKey {
	switch event.Name() {
	case "Enter":
		n := c.fileTree.GetCurrentNode().GetText()
		for _, f := range c.commit.Files {
			if n == f.Filename {
				c.patch.SetText(f.Patch)
				return event
			}
		}
	}

	return event
}

func (c *CommitPage) fetchCommit(repo string, sha string) error {
	_, err := FetchJson(fmt.Sprintf("https://api.github.com/repos/%s/commits/%s", repo, sha), &c.commit)
	if err != nil {
		return err
	}
	return nil
}
