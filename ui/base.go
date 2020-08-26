package ui

import (
	"github.com/eihigh/goban"
	"github.com/gdamore/tcell"
	"github.com/lusingander/ecr-browser/domain"
	"github.com/lusingander/ecr-browser/layout"
	"github.com/lusingander/ecr-browser/util"
)

const (
	mainViewTitle = "ECR BROWSER"
)

var (
	breadcrumbBases = []string{"ECR", "REPOSITORIES"}
)

type baseView struct {
	base *goban.Box
	*layout.Breadcrumb
	*gridLayout
	current *defaultView
	repo    *defaultView
	es      goban.Events
	focused operator
}

type operator interface {
	operate(*tcell.EventKey)
}

func (v *baseView) operate(key *tcell.EventKey) {
	if v.focused != nil {
		v.focused.operate(key)
	}
}

type defaultView struct {
	list   *listViewBase
	detail detailView
}

func newBaseView(cli domain.ContainerClient, es goban.Events) (*baseView, error) {
	b := goban.Screen()
	g := createGrid(util.InsideSides(b, 1, 2, 1, 1))
	dv, err := newRepositoryDefaultView(g, cli)
	if err != nil {
		return nil, err
	}
	bc := newECRBreadcrumb(b.Pos.X+2, b.Pos.Y+1, b.Size.X-3)
	bv := createBaseView(b, bc, g, dv, es)
	util.PushViews(bv, bc, dv.list, dv.detail)
	return bv, nil
}

func createBaseView(b *goban.Box, bc *layout.Breadcrumb, g *gridLayout, dv *defaultView, es goban.Events) *baseView {
	bv := &baseView{
		base:       b,
		Breadcrumb: bc,
		gridLayout: g,
		current:    dv,
		repo:       dv,
		es:         es,
		focused:    dv.list,
	}
	dv.list.setParent(bv)
	return bv
}

func newRepositoryDefaultView(g *gridLayout, cli domain.ContainerClient) (*defaultView, error) {
	lv, err := newRepositoryListView(g.list, cli)
	if err != nil {
		return nil, err
	}
	dv := newRepositoryDetailView(g.detail)
	lv.addObserver(dv)
	return &defaultView{lv.listViewBase, dv}, nil
}

func (v *baseView) View() {
	v.base.Enclose(mainViewTitle)
}

func newECRBreadcrumb(x, y, w int) *layout.Breadcrumb {
	b := layout.NewBreadcrumb(x, y, w)
	for _, v := range breadcrumbBases {
		b.Push(v)
	}
	return b
}

func (v *baseView) pushBreadcrumb(s string) {
	v.Breadcrumb.Push(s)
}

func (v *baseView) popBreadcrumb() string {
	return v.Breadcrumb.Pop()
}

func (v *baseView) displayRepositoryView() {
	v.updateBaseViews(v.repo)
	v.popBreadcrumb()
}

func (v *baseView) displayImageViews(repo string) error {
	loading := layout.NewLoadingDialog(v.base, v.es)
	go loading.Display()
	defer loading.Close()

	dv, err := v.loadImageDefaultView(v.gridLayout, client, repo)
	if err != nil {
		return err
	}
	v.updateBaseViews(dv)
	v.pushBreadcrumb(repo)
	return nil
}

func (v *baseView) loadImageDefaultView(g *gridLayout, cli domain.ContainerClient, repo string) (*defaultView, error) {
	return v.newImageDefaultView(g, cli, repo)
}

func (v *baseView) newImageDefaultView(g *gridLayout, cli domain.ContainerClient, repo string) (*defaultView, error) {
	lv, err := newImageListView(g.list, cli, repo)
	if err != nil {
		return nil, err
	}
	dv := newImageDetailView(g.detail)
	lv.addObserver(dv)
	ret := &defaultView{lv.listViewBase, dv}
	return ret, nil
}

func (v *baseView) updateBaseViews(dv *defaultView) {
	util.RemoveViews(dv.list, dv.detail)
	v.current = dv
	util.PushViews(dv.list, dv.detail)
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

type detailView interface {
	goban.View
}
