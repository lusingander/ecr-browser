package main

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/eihigh/goban"
	"github.com/mattn/go-runewidth"
)

const (
	targetRegion = endpoints.ApNortheast1RegionID

	datetimeFormat = "2006-01-02 15:04:05"
)

const (
	countFormat = " %*d/%*d "
)

var (
	grid = goban.NewGrid(
		"    1fr    2fr",
		"1fr list  detail",
	)
)

func app(_ context.Context, es goban.Events) error {
	svc := createClient()

	mainView := newMainView()
	goban.PushView(mainView)

	rlView, err := newRepositoryListView(mainView.grid.list, svc)
	if err != nil {
		return err
	}
	rdView := newRepositoryDetailView(mainView.grid.detail)
	rlView.addObserver(rdView)
	goban.PushView(rlView)
	goban.PushView(rdView)

	var currentListView listView = rlView
	var currentDetailView detailView = rdView

	for {
		goban.Show()

		k := es.ReadKey()
		switch k.Rune() {
		case 'k':
			currentListView.selectPrev()
		case 'j':
			currentListView.selectNext()
		case 'g':
			currentListView.selectFirst()
		case 'G':
			currentListView.selectLast()
		case 'l':
			if rlView, ok := currentListView.(*repositoryListView); ok {
				goban.RemoveView(currentListView)
				goban.RemoveView(currentDetailView)
				// TODO: cache
				v, err := newImageListView(mainView.grid.list, svc, rlView.currentRepositoryName())
				if err != nil {
					return err
				}
				currentListView = v
				idv := newImageDetailView(mainView.grid.detail)
				v.addObserver(idv)
				currentDetailView = idv
				goban.PushView(currentListView)
				goban.PushView(currentDetailView)
			}
		case 'h':
			if _, ok := currentListView.(*imageListView); ok {
				goban.RemoveView(currentListView)
				goban.RemoveView(currentDetailView)
				currentListView = rlView
				currentDetailView = rdView
				goban.PushView(currentListView)
				goban.PushView(currentDetailView)
			}
		default:
			return nil
		}
	}
}

type mainView struct {
	base *goban.Box
	grid *gridLayout
}

func newMainView() *mainView {
	b := goban.Screen()
	g := createGrid(b.InsideSides(1, 1, 1, 1))
	return &mainView{b, g}
}

func (v *mainView) View() {
	v.base.Enclose("ECR BROWSER")
}

type gridLayout struct {
	list   *goban.Box
	detail *goban.Box
}

func createGrid(b *goban.Box) *gridLayout {
	list := b.GridItem(grid, "list")
	detail := b.GridItem(grid, "detail")
	return &gridLayout{list, detail}
}

type listView interface {
	goban.View
	selectNext()
	selectPrev()
	selectFirst()
	selectLast()
}

type repositoryObserver interface {
	update(r *repository)
}

type imageListView struct {
	cur       int
	box       *goban.Box
	images    []*image
	observers []imageObserver
}

func newImageListView(b *goban.Box, svc *ecr.ECR, repoName string) (*imageListView, error) {
	imgs, err := fetchImages(svc, repoName)
	if err != nil {
		return nil, err
	}
	return &imageListView{
		box:    b,
		images: imgs,
	}, nil
}

func (v *imageListView) View() {
	// TODO: scroll / paging
	b := v.box.Enclose("IMAGE LIST")
	for i, img := range v.images {
		if v.cur == i {
			b.Print("> ")
		} else {
			b.Print("  ")
		}
		b.Puts(img.getTag())
	}
	createFooter(v.box, v).Print(currentCountStr(v))
}

func (v *imageListView) length() int {
	return len(v.images)
}

func (v *imageListView) cursor() int {
	return v.cur
}

func (v *imageListView) selectNext() {
	if v != nil && v.cur < len(v.images)-1 {
		v.cur++
		v.notify()
	}
}

func (v *imageListView) selectPrev() {
	if v != nil && v.cur > 0 {
		v.cur--
		v.notify()
	}
}

func (v *imageListView) selectFirst() {
	if v != nil && v.cur > 0 {
		v.cur = 0
		v.notify()
	}
}

func (v *imageListView) selectLast() {
	if v != nil && v.cur < len(v.images)-1 {
		v.cur = len(v.images) - 1
		v.notify()
	}
}

func (v *imageListView) notify() {
	current := v.current()
	for _, o := range v.observers {
		o.update(current)
	}
}

func (v *imageListView) current() *image {
	return v.images[v.cur]
}

func (v *imageListView) addObserver(o imageObserver) {
	v.observers = append(v.observers, o)
	o.update(v.current())
}

type imageObserver interface {
	update(i *image)
}

type repositoryListView struct {
	cur          int
	box          *goban.Box
	repositories []*repository
	observers    []repositoryObserver
}

func newRepositoryListView(b *goban.Box, svc *ecr.ECR) (*repositoryListView, error) {
	repos, err := fetchRepositories(svc)
	if err != nil {
		return nil, err
	}
	// TODO: sort
	return &repositoryListView{
		box:          b,
		repositories: repos,
	}, nil
}

func (v *repositoryListView) View() {
	// TODO: scroll / paging
	b := v.box.Enclose("REPOSITORY LIST")
	for i, r := range v.repositories {
		if v.cursor() == i {
			b.Print("> ")
		} else {
			b.Print("  ")
		}
		b.Puts(r.name)
	}
	createFooter(v.box, v).Print(currentCountStr(v))
}

func (v *repositoryListView) length() int {
	return len(v.repositories)
}

func (v *repositoryListView) cursor() int {
	return v.cur
}

func createFooter(b *goban.Box, c cursorer) *goban.Box {
	l := calcCountStrMaxLen(c)
	h := 1
	y := b.Pos.Y + b.Size.Y - h
	x := b.Pos.X + b.Size.X - l - 1 // right justify
	w := l
	return goban.NewBox(x, y, w, h)
}

type cursorer interface {
	length() int
	cursor() int
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

func (v *repositoryListView) selectNext() {
	if v.cur < len(v.repositories)-1 {
		v.cur++
		v.notify()
	}
}

func (v *repositoryListView) selectPrev() {
	if v.cur > 0 {
		v.cur--
		v.notify()
	}
}

func (v *repositoryListView) selectFirst() {
	if v.cur > 0 {
		v.cur = 0
		v.notify()
	}
}

func (v *repositoryListView) selectLast() {
	if v.cur < len(v.repositories)-1 {
		v.cur = len(v.repositories) - 1
		v.notify()
	}
}

func (v *repositoryListView) notify() {
	current := v.current()
	for _, o := range v.observers {
		o.update(current)
	}
}

func (v *repositoryListView) current() *repository {
	return v.repositories[v.cur]
}

func (v *repositoryListView) currentRepositoryName() string {
	return v.current().name
}

func (v *repositoryListView) addObserver(o repositoryObserver) {
	v.observers = append(v.observers, o)
	o.update(v.current())
}

func setting() {
	runewidth.DefaultCondition = &runewidth.Condition{EastAsianWidth: false}
}

func main() {
	setting()
	goban.Main(app)
}
