package ui

import (
	"context"
	"fmt"

	"github.com/eihigh/goban"
	"github.com/gdamore/tcell"
	"github.com/lusingander/ecr-browser/domain"
	"github.com/mattn/go-runewidth"
)

const (
	gridAreaList   = "list"
	gridAreaDetail = "detail"
)

var (
	client domain.ContainerClient
)

var (
	grid = goban.NewGrid(
		fmt.Sprintf("    1fr  2fr"),
		fmt.Sprintf("1fr  %s   %s", gridAreaList, gridAreaDetail),
	)
)

func app(_ context.Context, es goban.Events) error {
	bv, err := newBaseView(client, es)
	if err != nil {
		return err
	}

	for {
		goban.Show()
		switch key := es.ReadKey(); key.Rune() {
		case 'k':
			bv.current.list.selectPrev()
		case 'j':
			bv.current.list.selectNext()
		case 'g':
			bv.current.list.selectFirst()
		case 'G':
			bv.current.list.selectLast()
		case 'l':
			bv.displayImageViews(client)
		case 'h':
			bv.displayRepositoryView()
		case 'o':
			bv.openWebBrowser()
		case 'q':
			return nil // quit
		default:
			switch key.Key() {
			case tcell.KeyCtrlC:
				return nil // quit
			}
		}
	}
}

func setting() {
	runewidth.DefaultCondition = &runewidth.Condition{EastAsianWidth: false}
}

func Start(cli domain.ContainerClient) error {
	client = cli
	setting()
	return goban.Main(app)
}
