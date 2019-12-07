package main

import (
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/eihigh/goban"
)

const (
	mainViewTitle  = "ECR BROWSER"
	breadcrumbBase = "ECR > REPOSITORIES"
)

type baseView struct {
	base *goban.Box
	*gridLayout
	current *defaultView
	repo    *defaultView
	images  cacheMap
}

type cacheMap map[string]*defaultView

type defaultView struct {
	list   listView
	detail detailView
}

func newBaseView(svc *ecr.ECR) (*baseView, error) {
	b := goban.Screen()
	g := createGrid(insideSides(b, 1, 2, 1, 1))
	dv, err := newRepositoryDefaultView(g, svc)
	if err != nil {
		return nil, err
	}
	bv := &baseView{
		base:       b,
		gridLayout: g,
		current:    dv,
		repo:       dv,
		images:     make(cacheMap),
	}
	pushViews(bv, dv.list, dv.detail)
	return bv, nil
}

func newRepositoryDefaultView(g *gridLayout, svc *ecr.ECR) (*defaultView, error) {
	lv, err := newRepositoryListView(g.list, svc)
	if err != nil {
		return nil, err
	}
	dv := newRepositoryDetailView(g.detail)
	lv.addObserver(dv)
	return &defaultView{lv, dv}, nil
}

func (v *baseView) View() {
	v.base.Enclose(mainViewTitle)
	v.printBreadcrumb()
}

func (v *baseView) printBreadcrumb() {
	repo, ok := v.getParentRepositoryName()
	bc := goban.NewBox(v.base.Pos.X+2, v.base.Pos.Y+1, v.base.Size.X-3, 1)
	if ok {
		bc.Puts(breadcrumbBase + " > " + repo)
	} else {
		bc.Puts(breadcrumbBase)
	}
}

func (v *baseView) displayRepositoryView() {
	if _, ok := v.current.list.(*imageListView); ok {
		v.updateBaseViews(v.repo)
	}
}

func (v *baseView) displayImageViews(svc *ecr.ECR) error {
	repo, ok := v.getCurrentRepositoryName()
	if !ok {
		return nil
	}
	dv, err := v.newImageDefaultView(v.gridLayout, svc, repo)
	if err != nil {
		return err
	}
	v.updateBaseViews(dv)
	return nil
}

func (v *baseView) loadImageDefaultView(g *gridLayout, svc *ecr.ECR, repo string) (*defaultView, error) {
	if i, ok := v.images[repo]; ok {
		return i, nil
	}
	return v.newImageDefaultView(g, svc, repo)
}

func (v *baseView) newImageDefaultView(g *gridLayout, svc *ecr.ECR, repo string) (*defaultView, error) {
	lv, err := newImageListView(g.list, svc, repo)
	if err != nil {
		return nil, err
	}
	dv := newImageDetailView(g.detail)
	lv.addObserver(dv)
	ret := &defaultView{lv, dv}
	v.images[repo] = ret
	return ret, nil
}

func (v *baseView) getCurrentRepositoryName() (name string, ok bool) {
	lv, ok := v.current.list.(*repositoryListView)
	if !ok {
		return "", false
	}
	return lv.currentRepositoryName(), true
}

func (v *baseView) getParentRepositoryName() (name string, ok bool) {
	lv, ok := v.current.list.(*imageListView)
	if !ok {
		return "", false
	}
	return lv.repository, true
}

func (v *baseView) updateBaseViews(dv *defaultView) {
	removeViews(dv.list, dv.detail)
	v.current = dv
	pushViews(dv.list, dv.detail)
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
	cursor() int // return -1 if length is zero
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
