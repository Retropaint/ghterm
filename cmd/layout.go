package cmd

import (
	"github.com/rivo/tview"
)

type LayoutStruct struct {
	App        *tview.Application
	pages      *tview.Pages
	searchPage SearchPage
	homePage   *tview.Flex
}

var (
	Layout LayoutStruct
)

func (l *LayoutStruct) Refresh() {
	l.App.Draw()
}

func (l *LayoutStruct) Run() {
	if err := l.App.SetRoot(l.pages, true).Run(); err != nil {
		panic(err)
	}
}

func Init() {
	Layout.App = tview.NewApplication()

	Layout.searchPage.Init()

	Layout.pages = tview.NewPages()
	Layout.pages.AddPage("search", Layout.searchPage, true, true)
}
