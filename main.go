package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/eihigh/goban"
	"github.com/mattn/go-runewidth"
)

const (
	targetRegion = endpoints.ApNortheast1RegionID
)

const (
	datetimeFormat = "2006-01-02 15:04:05"
	countFormat    = " %*d/%*d "
)

const (
	gridAreaList   = "list"
	gridAreaDetail = "detail"
)

var (
	grid = goban.NewGrid(
		fmt.Sprintf("    1fr  2fr"),
		fmt.Sprintf("1fr  %s   %s", gridAreaList, gridAreaDetail),
	)
)

func app(_ context.Context, es goban.Events) error {
	svc := createClient()
	bv, err := newBaseView(svc)
	if err != nil {
		return err
	}

	for {
		goban.Show()
		switch es.ReadKey().Rune() {
		case 'k':
			bv.currentListView.selectPrev()
		case 'j':
			bv.currentListView.selectNext()
		case 'g':
			bv.currentListView.selectFirst()
		case 'G':
			bv.currentListView.selectLast()
		case 'l':
			bv.displayImageViews(svc)
		case 'h':
			bv.displayRepositoryView()
		default:
			return nil
		}
	}
}

func setting() {
	runewidth.DefaultCondition = &runewidth.Condition{EastAsianWidth: false}
}

func main() {
	setting()
	goban.Main(app)
}
