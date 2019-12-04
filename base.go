package main

import (
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/eihigh/goban"
)

type baseView struct {
	base *goban.Box
	*gridLayout
	currentListView   listView
	currentDetailView detailView
	repoListView      listView
	repoDetailView    detailView
}

func newBaseView(svc *ecr.ECR) (*baseView, error) {
	b := goban.Screen()
	g := createGrid(insideSides(b, 1, 2, 1, 1))
	lv, dv, err := newRepositoryDefaultView(g, svc)
	if err != nil {
		return nil, err
	}
	bv := &baseView{
		base:              b,
		gridLayout:        g,
		currentListView:   lv,
		currentDetailView: dv,
		repoListView:      lv,
		repoDetailView:    dv,
	}
	pushViews(bv, lv, dv)
	return bv, nil
}

func newRepositoryDefaultView(g *gridLayout, svc *ecr.ECR) (listView, detailView, error) {
	lv, err := newRepositoryListView(g.list, svc)
	if err != nil {
		return nil, nil, err
	}
	dv := newRepositoryDetailView(g.detail)
	lv.addObserver(dv)
	return lv, dv, nil
}

func (v *baseView) View() {
	v.base.Enclose("ECR BROWSER")
}

func (v *baseView) displayRepositoryView() {
	if _, ok := v.currentListView.(*imageListView); ok {
		v.updateBaseViews(v.repoListView, v.repoDetailView)
	}
}

func (v *baseView) displayImageViews(svc *ecr.ECR) error {
	repo, ok := v.getCurrentRepositoryName()
	if !ok {
		return nil
	}
	ilv, idv, err := newImageDefaultView(v.gridLayout, svc, repo)
	if err != nil {
		return err
	}
	v.updateBaseViews(ilv, idv)
	return nil
}

func newImageDefaultView(g *gridLayout, svc *ecr.ECR, repo string) (listView, detailView, error) {
	lv, err := newImageListView(g.list, svc, repo)
	if err != nil {
		return nil, nil, err
	}
	dv := newImageDetailView(g.detail)
	lv.addObserver(dv)
	return lv, dv, nil
}

func (v *baseView) getCurrentRepositoryName() (name string, ok bool) {
	rlv, ok := v.currentListView.(*repositoryListView)
	if !ok {
		return "", false
	}
	return rlv.currentRepositoryName(), true
}

func (v *baseView) updateBaseViews(newLv listView, newDv detailView) {
	removeViews(v.currentListView, v.currentDetailView)
	v.currentListView = newLv
	v.currentDetailView = newDv
	pushViews(newLv, newDv)
}

type gridLayout struct {
	list   *goban.Box
	detail *goban.Box
}

func createGrid(b *goban.Box) *gridLayout {
	list := b.GridItem(grid, gridAreaList)
	detail := b.GridItem(grid, gridAreaDetail)
	return &gridLayout{list, detail}
}

type cursorer interface {
	length() int
	cursor() int
}

func createFooter(b *goban.Box, c cursorer) *goban.Box {
	l := calcCountStrMaxLen(c)
	h := 1
	y := b.Pos.Y + b.Size.Y - h
	x := b.Pos.X + b.Size.X - l - 1 // right justify
	w := l
	return goban.NewBox(x, y, w, h)
}

func calcCountStrMaxLen(c cursorer) int {
	return len(countStr(c, c.length()))
}

func currentCountStr(c cursorer) string {
	return countStr(c, c.cursor()+1)
}

func countStr(c cursorer, n int) string {
	l := c.length()
	d := len(strconv.Itoa(l))
	return fmt.Sprintf(countFormat, d, n, d, l)
}

type listView interface {
	goban.View
	selectNext()
	selectPrev()
	selectFirst()
	selectLast()
}

type detailView interface {
	goban.View
}
