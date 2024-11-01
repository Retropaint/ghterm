package cmd

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Accounts for all results (repos, users, etc).
type SearchResult struct {
	Total_count int
	Items       []Repo
}

type Repo struct {
	Name        string
	Full_name   string
	Description string
}

type SearchPage struct {
	*tview.Flex
	input *tview.InputField
	list  *tview.TextView

	result SearchResult
	wg     sync.WaitGroup
}

func (sp *SearchPage) Init() {
	sp.input = tview.NewInputField()
	sp.input.SetFieldBackgroundColor(tcell.ColorBlack)
	sp.input.SetBorder(true)
	sp.input.SetDoneFunc(sp.search)

	sp.list = tview.NewTextView()
	sp.list.SetBorder(true)
	sp.list.SetRegions(true)
	sp.list.SetDynamicColors(true)

	sp.Flex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(sp.input, 3, 0, true).
		AddItem(sp.list, 0, 3, false)

	sp.clearList("Use the search above to find a repository!")
}

func (sp *SearchPage) search(key tcell.Key) {
	sp.clearList("loading...")
	go func() {
		sp.wg.Add(1)
		sp.fetch(&sp.result, "https://api.github.com/search/repositories?q="+sp.input.GetText()+"&per_page=3")
		sp.clearList("")
		sp.populateList()
	}()
}

func (sp *SearchPage) clearList(startingText string) {
	sp.list.SetText(startingText)
}

func (sp *SearchPage) populateList() {
	for _, r := range sp.result.Items {
		// Since slash isn't accepted for regions, replace it with dot
		region := strings.Replace(r.Full_name, "/", ".", 1)

		fmt.Fprintln(sp.list, "[\""+region+"\"][white]"+r.Name+"[-][lightgrey] - "+r.Description+`[-][""]`)
	}
}

func (sp *SearchPage) fetch(obj any, url string) {
	defer sp.wg.Done()
	client := &http.Client{}

	request, err := http.NewRequest("GET", url, nil)
	if err != nil {
	}

	response, err := client.Do(request)
	if err != nil {
	}

	err = json.NewDecoder(response.Body).Decode(obj)
}
