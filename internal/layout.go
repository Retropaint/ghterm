package internal

import (
	"encoding/json"
	"net/http"

	"github.com/Retropaint/ghterm/internal/config"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type LayoutStruct struct {
	App         *tview.Application
	Pages       *tview.Pages
	searchPage  SearchPage
	repoPage    RepoPage
	commitsPage CommitsPage
	commitPage  CommitPage
	homePage    *tview.Flex
	pageHistory []string
}

var (
	Layout LayoutStruct
	Client http.Client
)

func (l *LayoutStruct) Run() {
	if err := l.App.SetRoot(l.Pages, true).Run(); err != nil {
		panic(err)
	}
}

// Basic fetch that returns a response. Uses the global client.
func Fetch(url string) (*http.Response, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	//request.Header.Set("Authorization", "ghp_zCgtRbrhrrREbLDTOtWiG5yDbrk3gJ4HUga")

	response, err := Client.Do(request)
	if err != nil {
		return nil, err
	}

	return response, nil
}

// Just like Fetch(), but decodes JSON.
func FetchJson(url string, obj any) (*http.Response, error) {
	response, err := Fetch(url)
	if err != nil {
		return response, err
	}

	json.NewDecoder(response.Body).Decode(obj)

	return response, nil
}

func Init() {
	err := config.LoadCfg()
	if err != nil {
		panic(err)
	}

	Layout.App = tview.NewApplication()
	Layout.App.SetInputCapture(Layout.onInputCapture)

	Layout.searchPage.Init()
	Layout.repoPage.Init()
	Layout.commitsPage.Init()
	Layout.commitPage.Init()

	Layout.Pages = tview.NewPages()
	Layout.Pages.AddPage("search", Layout.searchPage, true, true)
	Layout.Pages.AddPage("repo", Layout.repoPage, true, false)
	Layout.Pages.AddPage("commits", Layout.commitsPage, true, false)
	Layout.Pages.AddPage("commit", Layout.commitPage, true, false)
}

func OpenRepo(repo string) {
	SwitchToPage("repo")
	Layout.repoPage.Reset()
	Layout.repoPage.GetRepo(repo)
}

func OpenCommits(repo string) {
	SwitchToPage("commits")
	Layout.repoPage.Reset()
	Layout.commitsPage.GetCommits(repo)
}

func OpenCommit(repo string, sha string) {
	SwitchToPage("commit")
	Layout.commitPage.Reset()
	Layout.commitPage.OpenCommit(repo, sha, "")
}

func (l *LayoutStruct) onInputCapture(event *tcell.EventKey) *tcell.EventKey {
	switch event.Name() {
	case "Backspace2":
		if len(l.pageHistory) == 1 {
			return event
		}
		SwitchToPage(l.pageHistory[len(l.pageHistory)-2])
		l.pageHistory = l.pageHistory[:len(l.pageHistory)-2]
	}
	return event
}

func SwitchToPage(page string) {
	Layout.Pages.SwitchToPage(page)
	Layout.pageHistory = append(Layout.pageHistory, page)
}
