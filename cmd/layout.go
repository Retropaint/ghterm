package cmd

import (
	"net/http"

	"github.com/rivo/tview"
)

type LayoutStruct struct {
	App        *tview.Application
	Pages      *tview.Pages
	searchPage SearchPage
	homePage   *tview.Flex
	RepoPage   RepoPage
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

func Fetch(url string) (*http.Response, error) {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	response, err := Client.Do(request)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func Init() {
	Layout.App = tview.NewApplication()

	Layout.searchPage.Init()
	Layout.RepoPage.Init()

	Layout.Pages = tview.NewPages()
	Layout.Pages.AddPage("search", Layout.searchPage, true, true)
	Layout.Pages.AddPage("repo", Layout.RepoPage, true, false)
}

func OpenRepo(repo string) {
	Layout.Pages.SwitchToPage("rupo")
	Layout.RepoPage.GetRepo(repo)
}
