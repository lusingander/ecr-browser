package ui

import (
	"github.com/eihigh/goban"
	"github.com/gdamore/tcell"
	"github.com/lusingander/ecr-browser/layout"
	"github.com/lusingander/ecr-browser/util"
)

const (
	mainViewTitle = "ECR BROWSER"
)

var (
	breadcrumbBases = []string{"ECR", "REPOSITORIES"}
)

type operator interface {
	operate(*tcell.EventKey)
}

type viewStack [][]goban.View

func newViewStack() viewStack {
	return make([][]goban.View, 0)
}

func (s viewStack) push(vs ...goban.View) {
	s = append(s, vs)
}

func (s viewStack) pop(vs ...goban.View) []goban.View {
	ret := s[len(s)-1]
	s = s[:len(s)-1]
	return ret
}

type ui struct {
	*baseView
	viewStack
	focused operator
}

func newUI(es goban.Events) (*ui, error) {
	ui := &ui{}
	baseView, err := newBaseView(es)
	if err != nil {
		return nil, err
	}
	ui.baseView = baseView
	ui.viewStack = newViewStack()
	ui.loadRepositoryView(true)
	return ui, nil
}

func (u *ui) pushViews(vs ...goban.View) {
	u.viewStack.push(vs...)
	util.PushViews(vs...)
}

func (u *ui) popViews() {
	if len(u.viewStack) > 0 {
		vs := u.viewStack.pop()
		util.RemoveViews(vs...)
	}
}

func (u *ui) operate(key *tcell.EventKey) {
	if u.focused != nil {
		u.focused.operate(key)
	}
}

func (u *ui) loadRepositoryView(init bool) error {
	loading := layout.NewLoadingDialog(u.baseView.base, u.baseView.es)
	go loading.Display()
	defer loading.Close()

	lv, dv, err := u.baseView.newRepositoryView()
	if err != nil {
		return err
	}
	lv.setBaseUI(u)
	u.pushViews(lv, dv)
	u.focused = lv
	if !init {
		u.baseView.popBreadcrumb()
	}
	return nil
}

func (u *ui) loadImageViews(repo string) error {
	loading := layout.NewLoadingDialog(u.baseView.base, u.baseView.es)
	go loading.Display()
	defer loading.Close()

	lv, dv, err := u.baseView.newImageView(repo)
	if err != nil {
		return err
	}
	lv.setBaseUI(u)
	u.popViews()
	u.pushViews(lv, dv)
	u.focused = lv
	u.baseView.pushBreadcrumb(repo)
	return nil
}

type baseView struct {
	base *goban.Box
	*layout.Breadcrumb
	*gridLayout
	es goban.Events
}

func newBaseView(es goban.Events) (*baseView, error) {
	bv := &baseView{es: es}
	b := goban.Screen()
	bv.base = b
	bv.createGrid(util.InsideSides(b, 1, 2, 1, 1))
	bv.newECRBreadcrumb(b.Pos.X+2, b.Pos.Y+1, b.Size.X-3)
	util.PushViews(bv, bv.Breadcrumb)
	return bv, nil
}

func (v *baseView) View() {
	v.base.Enclose(mainViewTitle)
}

func (v *baseView) newECRBreadcrumb(x, y, w int) {
	b := layout.NewBreadcrumb(x, y, w)
	for _, v := range breadcrumbBases {
		b.Push(v)
	}
	v.Breadcrumb = b
}

func (v *baseView) pushBreadcrumb(s string) {
	v.Breadcrumb.Push(s)
}

func (v *baseView) popBreadcrumb() string {
	return v.Breadcrumb.Pop()
}

func (v *baseView) newRepositoryView() (*repositoryListView, *repositoryDetailView, error) {
	lv, err := newRepositoryListView(v.gridLayout.list)
	if err != nil {
		return nil, nil, err
	}
	dv := newRepositoryDetailView(v.gridLayout.detail)
	lv.addObserver(dv)
	return lv, dv, nil
}

func (v *baseView) newImageView(repo string) (*imageListView, *imageDetailView, error) {
	lv, err := newImageListView(v.gridLayout.list, repo)
	if err != nil {
		return nil, nil, err
	}
	dv := newImageDetailView(v.gridLayout.detail)
	lv.addObserver(dv)
	return lv, dv, nil
}

type gridLayout struct {
	list   *goban.Box
	detail *goban.Box
}

func (v *baseView) createGrid(b *goban.Box) {
	list := b.GridItem(grid, gridAreaList)
	detail := b.GridItem(grid, gridAreaDetail)
	v.gridLayout = &gridLayout{list, detail}
}
