package internal

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"strings"

	"github.com/Retropaint/ghterm/internal/config"
	"github.com/alecthomas/chroma/v2/quick"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type RepoPage struct {
	*tview.Flex
	fileView     *tview.TextView
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
	focusedFile    *RepoContent
}

type RepoContent struct {
	Name     string
	Path     string
	Type     string
	data     string
	children []*RepoContent
}

func (rp *RepoPage) Init() {
	rp.fileView = tview.NewTextView()
	rp.fileView.SetBorder(true)
	rp.fileView.SetTitle("README.md")
	rp.fileView.SetTitleAlign(tview.AlignLeft)
	rp.fileView.SetText("Loading...")
	rp.fileView.SetDynamicColors(true)

	rp.fileTree = tview.NewTreeView()
	rp.fileTree.SetBorder(true)
	rp.fileTree.SetTitle("Files")
	rp.fileTree.SetTitleAlign(tview.AlignLeft)

	rp.Flex = tview.NewFlex().
		SetDirection(tview.FlexColumn).
		AddItem(rp.fileTree, 0, 1, false).
		AddItem(rp.fileView, 0, 1, false)

	rp.Flex.SetInputCapture(rp.onInputCapture)
	rp.fileTree.SetInputCapture(rp.fileTreeOnInputCapture)
	rp.fileView.SetInputCapture(rp.fileViewOnInputCapture)
}

func (rp *RepoPage) onInputCapture(event *tcell.EventKey) *tcell.EventKey {
	if rp.repo404 {
		Layout.Pages.SwitchToPage("search")
	}
	switch event.Name() {
	case "Ctrl+F":
		Layout.App.SetFocus(rp.fileTree)
	case "Ctrl+L":
		Layout.App.SetFocus(rp.fileView)
	}

	return event
}

func (rp *RepoPage) fileTreeOnInputCapture(event *tcell.EventKey) *tcell.EventKey {
	switch event.Name() {
	case "Enter":
		user := strings.Split(rp.repo.Full_name, "/")[0]
		repo := strings.Split(rp.repo.Full_name, "/")[1]
		node := rp.fileTree.GetCurrentNode()

		nodePath := ""
		for i, node := range rp.fileTree.GetPath(node) {
			if i == 0 {
				continue
			} else if i > 1 {
				nodePath += "/"
			}

			nodePath += node.GetText()
		}

		for i, file := range rp.repo.contents {
			if nodePath == file.Path && file.Name == node.GetText() {
				rp.repo.focusedFile = &rp.repo.contents[i]
				break
			}
		}

		if rp.repo.focusedFile.Type == "dir" {
			rp.tryFolder(node, user, repo, rp.repo.focusedFile.Path)
			return event
		}

		if rp.repo.focusedFile.data != "" {
			rp.openFile(true)
			return event
		}

		rp.fileView.SetText("Loading...")
		rp.fileView.SetTitle(rp.repo.focusedFile.Name)
		go func() {
			rp.repo.focusedFile.data, _ = rp.fetchFile(user, repo, rp.repo.Default_branch, nodePath)
			rp.openFile(false)
		}()
	}
	return event
}

func (rp *RepoPage) fileViewOnInputCapture(event *tcell.EventKey) *tcell.EventKey {
	switch event.Name() {
	case "Enter":
		rp.openFileInEditor()
		return nil
	}
	return event
}

func (rp *RepoPage) openFile(goroutine bool) {
	rp.fileView.SetText("Loading...")
	rp.fileView.SetTitle(rp.repo.focusedFile.Name)
	if !goroutine {
		rp.outputFileToView()
		Layout.App.Draw()
		return
	}

	go func() {
		rp.outputFileToView()
		Layout.App.Draw()
	}()
}

// outTypes:
// 0 - basic
// 1 - Chroma
// 2 - Viewer of choice (via VIEWER env)
func (rp *RepoPage) outputFileToView() error {
	rp.fileView.SetText("")
	ansi := tview.ANSIWriter(rp.fileView)

	outType := 1
	if config.Cfg.CodeViewer != "" {
		outType = 2
	}

	switch outType {
	case 0:
		rp.fileView.SetText(rp.repo.focusedFile.data)
		return nil
	case 1:
		err := quick.Highlight(ansi, rp.repo.focusedFile.data, rp.repo.focusedFile.ext(), "terminal", "")
		if err != nil {
			return err
		}
		return nil
	case 2:
		temp, err := os.CreateTemp("", "_*."+rp.repo.focusedFile.ext())
		if err != nil {
			return err
		}
		temp.WriteString(rp.repo.focusedFile.data)
		temp.Close()

		var cmd *exec.Cmd
		if config.Cfg.CodeViewerParams != "" {
			cmd = exec.Command(config.Cfg.CodeViewer, config.Cfg.CodeViewerParams, temp.Name())
		} else {
			cmd = exec.Command(config.Cfg.CodeViewer, temp.Name())
		}
		cmd.Stdin = os.Stdin
		cmd.Stdout = ansi
		cmd.Stderr = os.Stderr
		cmd.Run()
	}

	return nil
}

func (rp *RepoPage) openFileInEditor() error {
	temp, err := os.CreateTemp("", "_*."+rp.repo.focusedFile.ext())
	if err != nil {
		return err
	}
	temp.WriteString(rp.repo.focusedFile.data)
	temp.Close()

	Layout.App.Suspend(func() {

		var cmd *exec.Cmd
		if config.Cfg.CodeEditorParams != "" {
			cmd = exec.Command(config.Cfg.CodeEditor, config.Cfg.CodeEditorParams, temp.Name())
		} else {
			cmd = exec.Command(config.Cfg.CodeEditor, temp.Name())
		}
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Run()
	})

	return nil
}

func (c *RepoContent) ext() string {
	s := strings.Split(c.Name, ".")
	return s[len(s)-1]
}

func (rp *RepoPage) tryFolder(node *tview.TreeNode, user string, repo string, filePath string) {
	if len(rp.repo.focusedFile.children) > 0 {
		if node.IsExpanded() {
			node.Collapse()
		} else {
			node.Expand()
		}
		return
	}
	originalName := node.GetText()
	node.SetText(node.GetText() + " (loading)")

	go func() {
		var c []RepoContent
		rp.fetchFolder(user, repo, rp.repo.Default_branch, filePath, &c)
		for _, file := range c {
			rp.repo.focusedFile.children = append(rp.repo.focusedFile.children, &file)
			new := tview.NewTreeNode(file.Name)
			node.AddChild(new)
			if file.Type == "dir" {
				new.SetColor(tcell.ColorPurple)
			}
			rp.repo.contents = append(rp.repo.contents, file)
		}

		node.SetText(originalName)
		Layout.App.Draw()
	}()

}

func (rp *RepoPage) GetRepo(name string) {
	user := strings.Split(name, "/")[0]
	repo := strings.Split(name, "/")[1]
	go func() {
		err := rp.fetchRepo(user, repo)
		if err != nil {
			rp.fileView.SetText(fmt.Sprint(err))
			rp.repo404 = true
			return
		}

		rp.fileTreeNode = tview.NewTreeNode("root")
		rp.fileTree.SetRoot(rp.fileTreeNode)
		rp.fileTree.SetCurrentNode(rp.fileTreeNode)

		// fetch the main readme.md of this repo
		for i, file := range rp.repo.contents {
			new := tview.NewTreeNode(file.Name)
			rp.fileTreeNode.AddChild(new)
			if file.Type == "dir" {
				new.SetColor(tcell.ColorPurple)
			}
			if strings.Index(strings.ToLower(file.Name), "readme.md") != -1 {
				rp.repo.contents[i].data, _ = rp.fetchFile(user, repo, rp.repo.Default_branch, file.Name)
				rp.repo.focusedFile = &rp.repo.contents[i]
			}
		}

		if len(rp.repo.contents) == 0 {
			rp.fileView.SetText("No files were loaded.")
		} else {
			rp.openFile(true)
		}

		Layout.App.Draw()
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

func (rp *RepoPage) fetchFolder(user string, repo string, branch string, folderName string, folder *[]RepoContent) error {
	response, err := FetchJson(fmt.Sprintf("https://api.github.com/repos/%s/%s/contents/%s?ref=%s", user, repo, folderName, branch), &folder)
	if err != nil {
		return err
	}
	if response.StatusCode == http.StatusNotFound {
		return errors.New("Repo doesn't exist. Press any button to return to the search page.")
	}

	return nil
}

func (rp *RepoPage) fetchFile(user string, repo string, branch string, file string) (string, error) {
	response, err := Fetch(fmt.Sprintf("https://raw.githubusercontent.com/%s/%s/refs/heads/%s/%s", user, repo, branch, file))
	if err != nil {
		return "", err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(response.Body)
	respBytes := buf.String()

	return string(respBytes), nil
}
