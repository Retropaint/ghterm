package cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// Accounts for all results (repos, users, etc).
type SearchResult struct {
	Total_count int
	Items       []Repo
}

type SearchPage struct {
	*tview.Flex
	input *tview.InputField
	list  *tview.TextView

	result    SearchResult
	repoIndex int
}

func (sp *SearchPage) Init() {
	sp.input = tview.NewInputField()
	sp.input.SetBorder(true)
	sp.input.SetDoneFunc(sp.search)
	sp.input.SetPlaceholder("search something...")
	sp.input.SetPlaceholderStyle(sp.input.GetPlaceholderStyle().
		Background(tcell.ColorBlack).
		Foreground(tcell.ColorGrey),
	)
	sp.input.SetFieldBackgroundColor(tcell.ColorBlack)

	sp.list = tview.NewTextView()
	sp.list.SetBorder(true)
	sp.list.SetRegions(true)
	sp.list.SetDynamicColors(true)
	sp.list.SetTitle("Results")
	sp.list.SetTitleAlign(tview.AlignLeft)

	sp.Flex = tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(sp.input, 3, 0, true).
		AddItem(sp.list, 0, 3, false)

	sp.Flex.SetInputCapture(sp.onInputCapture)
}

func (sp *SearchPage) search(key tcell.Key) {
	sp.clearList("[lightgrey]Loading \"" + sp.input.GetText() + "\"...[-]")
	go func() {
		err := sp.fetch(&sp.result, "https://api.github.com/search/repositories?q="+sp.input.GetText()+"&per_page=5")
		if err != nil {
			sp.clearList("An error occurred. Please ensure you have an Internet connection.")
			Layout.App.Draw()
		} else {
			sp.populateList()
			sp.repoIndex = 0
			if len(sp.result.Items) > 0 {
				sp.highlightResult()
			} else {
				sp.clearList("No results for \"" + sp.input.GetText() + "\".")
			}
			Layout.App.Draw()
		}
	}()
}

func (sp *SearchPage) clearList(startingText string) {
	sp.list.SetText(startingText)
}

func (sp *SearchPage) onInputCapture(event *tcell.EventKey) *tcell.EventKey {
	switch event.Name() {
	case "Ctrl+L":
		Layout.App.SetFocus(sp.list)
	case "Ctrl+F":
		Layout.App.SetFocus(sp.input)
	default:
		sp.listOnInputCapture(event)
	}
	return event
}

func (sp *SearchPage) listOnInputCapture(event *tcell.EventKey) {
	if Layout.App.GetFocus() != sp.list {
		return
	}
	switch event.Name() {
	case "Rune[j]":
		sp.repoIndex = min(len(sp.result.Items)-1, sp.repoIndex+1)
	case "Rune[k]":
		sp.repoIndex = max(0, sp.repoIndex-1)
	case "Enter":
		Layout.RepoPage.GetRepo(sp.result.Items[sp.repoIndex].Full_name)
		Layout.Pages.SwitchToPage("repo")
	default:
		Layout.App.Stop()
		fmt.Print(event.Name())
	}
	sp.highlightResult()
}

func (sp *SearchPage) highlightResult() {
	region := strings.Replace(sp.result.Items[sp.repoIndex].Full_name, "/", ".", 1)
	sp.list.Highlight(region)
}

func (sp *SearchPage) populateList() {
	sp.clearList("")
	for _, r := range sp.result.Items {
		// Since slash isn't accepted for regions, replace it with dot
		region := strings.Replace(r.Full_name, "/", ".", 1)

		desc := ""
		if r.Description != "" {
			desc += "[-][lightgrey] - " + r.Description + `[-][""]`
		}

		fmt.Fprintln(sp.list, "[\""+region+"\"][white]"+r.Name+desc)
	}
}

func (sp *SearchPage) fetch(obj any, url string) error {
	response, err := Fetch(url)
	if err != nil {
		return err
	}

	err = json.NewDecoder(response.Body).Decode(obj)
	if err != nil {
		return err
	}

	return nil
}
