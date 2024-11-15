package cmd

import (
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
	sp.list.SetInputCapture(sp.listOnInputCapture)
}

func (sp *SearchPage) search(key tcell.Key) {
	sp.clearList(fmt.Sprintf("[lightgrey]Loading \"%s\"...[-]", sp.input.GetText()))
	go func() {
		_, err := FetchJson(Url("https://api.github.com/search/repositories?q=\"%s\"&per_page=5", sp.input.GetText()), &sp.result)
		if err != nil {
			sp.clearList("An error occurred. Please ensure you have an Internet connection.")
			Layout.App.Draw()
		} else {
			sp.populateList()
			sp.repoIndex = 0
			if len(sp.result.Items) > 0 {
				sp.highlightResult()
			} else {
				sp.clearList(fmt.Sprintf("No results for \"%s\".", sp.input.GetText()))
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
	}
	return event
}

func (sp *SearchPage) listOnInputCapture(event *tcell.EventKey) *tcell.EventKey {
	switch event.Name() {
	case "Rune[j]":
		sp.repoIndex = min(len(sp.result.Items)-1, sp.repoIndex+1)
	case "Rune[k]":
		sp.repoIndex = max(0, sp.repoIndex-1)
	case "Enter":
		OpenRepo(sp.result.Items[sp.repoIndex].Full_name)
	}

	if len(sp.result.Items) > 0 {
		sp.highlightResult()
	}

	return event
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
			desc += fmt.Sprintf("[-][lightgrey] - %s [-][\"\"]", r.Description)
		}

		fmt.Fprintln(sp.list, fmt.Sprintf("[\"%s\"][white]%s%s", region, r.Name, desc))
	}
}
