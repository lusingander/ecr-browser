package main

import (
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/eihigh/goban"
)

const (
	mainViewTitle = "ECR BROWSER"
)

var (
	breadcrumbBases = []string{"ECR", "REPOSITORIES"}
)

type baseView struct {
	base *goban.Box
	*breadcrumb
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
	bc := newECRBreadcrumb(b.Pos.X+2, b.Pos.Y+1, b.Size.X-3)
	bv := &baseView{
		base:       b,
		breadcrumb: bc,
		gridLayout: g,
		current:    dv,
		repo:       dv,
		images:     make(cacheMap),
	}
	pushViews(bv, bc, dv.list, dv.detail)
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
}

func newECRBreadcrumb(x, y, w int) *breadcrumb {
	b := newBreadcrumb(x, y, w)
	for _, v := range breadcrumbBases {
		b.push(v)
	}
	return b
}

func (v *baseView) pushBreadcrumb(s string) {
	v.breadcrumb.push(s)
}

func (v *baseView) popBreadcrumb() string {
	return v.breadcrumb.pop()
}

func (v *baseView) displayRepositoryView() {
	if _, ok := v.current.list.(*imageListView); ok {
		v.updateBaseViews(v.repo)
		v.popBreadcrumb()
	}
}

func (v *baseView) displayImageViews(svc *ecr.ECR) error {
	repo, ok := v.getCurrentRepositoryName()
	if !ok {
		return nil
	}

	loading := newLoadingDialog(v.base)
	go loading.display()
	defer loading.close()

	dv, err := v.loadImageDefaultView(v.gridLayout, svc, repo)
	if err != nil {
		return err
	}
	v.updateBaseViews(dv)
	v.pushBreadcrumb(repo)
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
