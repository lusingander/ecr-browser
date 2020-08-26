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
	ui, err := newUI(es)
	if err != nil {
		return err
	}

	for {
		goban.Show()
		key := es.ReadKey()
		if key.Rune() == 'q' || key.Key() == tcell.KeyCtrlC {
			return nil
		}
		ui.operate(key)
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
