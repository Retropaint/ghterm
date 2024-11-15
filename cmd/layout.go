package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/rivo/tview"
)

type LayoutStruct struct {
	App        *tview.Application
	Pages      *tview.Pages
	searchPage SearchPage
	homePage   *tview.Flex
	repoPage   RepoPage
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

// Exactly like Sprintf. Since URLs are the lifeblood of ghterm, cleaner syntax for it goes a long way.
func Url(format string, a ...any) string {
	return fmt.Sprintf(format, a...)
}

func Init() {
	Layout.App = tview.NewApplication()

	Layout.searchPage.Init()
	Layout.repoPage.Init()

	Layout.Pages = tview.NewPages()
	Layout.Pages.AddPage("search", Layout.searchPage, true, true)
	Layout.Pages.AddPage("repo", Layout.repoPage, true, false)
}

func OpenRepo(repo string) {
	Layout.Pages.SwitchToPage("repo")
	Layout.repoPage.GetRepo(repo)
}
