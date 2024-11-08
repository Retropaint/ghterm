package cmd

import (
	"flag"
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

func Fetch(url string) *http.Response {
	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
	}

	response, err := Client.Do(request)
	if err != nil {
	}

	return response
}

func Init() {
	repo := flag.String("repo", "none", "Instantly opens a repo. \nExample: --repo retropaint/ghterm")
	flag.Parse()

	Layout.App = tview.NewApplication()

	Layout.searchPage.Init()
	Layout.RepoPage.Init(*repo)

	Layout.Pages = tview.NewPages()
	Layout.Pages.AddPage("search", Layout.searchPage, true, (*repo == "none"))
	Layout.Pages.AddPage("repo", Layout.RepoPage, true, (*repo != "none"))
}
