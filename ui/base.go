package ui

import (
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/eihigh/goban"
	"github.com/lusingander/ecr-browser/layout"
	"github.com/lusingander/ecr-browser/util"
	"github.com/pkg/browser"
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
	images  cacheMap
	es      goban.Events
}

type cacheMap map[string]*defaultView

type defaultView struct {
	list   listView
	detail detailView
}

func newBaseView(svc *ecr.ECR, es goban.Events) (*baseView, error) {
	b := goban.Screen()
	g := createGrid(util.InsideSides(b, 1, 2, 1, 1))
	dv, err := newRepositoryDefaultView(g, svc)
	if err != nil {
		return nil, err
	}
	bc := newECRBreadcrumb(b.Pos.X+2, b.Pos.Y+1, b.Size.X-3)
	bv := createBaseView(b, bc, g, dv, es)
	util.PushViews(bv, bc, dv.list, dv.detail)
	return bv, nil
}

func createBaseView(b *goban.Box, bc *layout.Breadcrumb, g *gridLayout, dv *defaultView, es goban.Events) *baseView {
	return &baseView{
		base:       b,
		Breadcrumb: bc,
		gridLayout: g,
		current:    dv,
		repo:       dv,
		images:     make(cacheMap),
		es:         es,
	}
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

	loading := layout.NewLoadingDialog(v.base, v.es)
	go loading.Display()
	defer loading.Close()

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
	util.RemoveViews(dv.list, dv.detail)
	v.current = dv
	util.PushViews(dv.list, dv.detail)
}

func (v *baseView) openWebBrowser() error {
	rv, ok := v.current.list.(*repositoryListView)
	if !ok {
		// TODO: error
		return nil
	}
	repo := rv.currentRepositoryName()
	url := createECRConsoleRepositoryURL(targetRegion, repo)
	return browser.OpenURL(url)
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
