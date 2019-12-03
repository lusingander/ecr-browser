package main

import (
	"context"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/dustin/go-humanize"
	"github.com/eihigh/goban"
	"github.com/mattn/go-runewidth"
)

const (
	targetRegion = endpoints.ApNortheast1RegionID

	datetimeFormat = "2006-01-02 15:04:05"
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
	View()
	selectNext()
	selectPrev()
	selectFirst()
	selectLast()
}

type detailView interface {
	View()
}

type repositoryDetailView struct {
	box      *goban.Box
	selected *repository
}

type repositoryObserver interface {
	update(r *repository)
}

func (v *repositoryDetailView) update(r *repository) {
	v.selected = r
}

func newRepositoryDetailView(b *goban.Box) *repositoryDetailView {
	return &repositoryDetailView{b, nil}
}

func (v *repositoryDetailView) View() {
	b := v.box.Enclose("DETAIL")
	if v.selected != nil {
		b.Puts("NAME:")
		b.Puts("  " + v.selected.name)
		b.Puts("URI:")
		b.Puts("  " + v.selected.uri)
		b.Puts("ARN:")
		b.Puts("  " + v.selected.arn)
		b.Puts("TAG MUTABILITY:")
		b.Puts("  " + v.selected.tagMutability)
		b.Puts("CREATED AT:")
		b.Puts("  " + v.selected.createdAtStr())
	}
}

type imageDetailView struct {
	box      *goban.Box
	selected *image
}

func (v *imageDetailView) update(i *image) {
	v.selected = i
}

func newImageDetailView(b *goban.Box) *imageDetailView {
	return &imageDetailView{b, nil}
}

func (v *imageDetailView) View() {
	b := v.box.Enclose("DETAIL")
	if v.selected != nil {
		b.Puts("TAGS:")
		for _, t := range v.selected.getTags() {
			b.Puts("  " + t)
		}
		b.Puts("PUSHED AT:")
		b.Puts("  " + v.selected.pushedAtStr())
		b.Puts("DIGEST:")
		b.Puts("  " + v.selected.digest)
		b.Puts("SIZE:")
		b.Puts("  " + v.selected.sizeStr())
		// TODO: show "selected / total count"
		// TODO: show current repository name
	}
}

type image struct {
	tags     []string
	pushedAt time.Time
	digest   string
	sizeByte int64
}

func newImage(i *ecr.ImageDetail) *image {
	return &image{
		tags:     aws.StringValueSlice(i.ImageTags),
		pushedAt: aws.TimeValue(i.ImagePushedAt),
		digest:   aws.StringValue(i.ImageDigest),
		sizeByte: aws.Int64Value(i.ImageSizeInBytes),
	}
}

func (i *image) getTag() string {
	return strings.Join(i.getTags(), ", ")
}

func (i *image) getTags() []string {
	if len(i.tags) == 0 {
		return []string{"<untagged>"}
	}
	return i.tags
}

func (i *image) pushedAtStr() string {
	// TODO: consider timezone
	return i.pushedAt.Format(datetimeFormat)
}

func (i *image) sizeStr() string {
	return humanize.Bytes(uint64(i.sizeByte))
}

type imageListView struct {
	cursor    int
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
	b := v.box.Enclose("IMAGE LIST")
	for i, img := range v.images {
		if v.cursor == i {
			b.Print("> ")
		} else {
			b.Print("  ")
		}
		b.Puts(img.getTag())
	}
}

func (v *imageListView) selectNext() {
	if v != nil && v.cursor < len(v.images)-1 {
		v.cursor++
		v.notify()
	}
}

func (v *imageListView) selectPrev() {
	if v != nil && v.cursor > 0 {
		v.cursor--
		v.notify()
	}
}

func (v *imageListView) selectFirst() {
	if v != nil && v.cursor > 0 {
		v.cursor = 0
		v.notify()
	}
}

func (v *imageListView) selectLast() {
	if v != nil && v.cursor < len(v.images)-1 {
		v.cursor = len(v.images) - 1
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
	return v.images[v.cursor]
}

func (v *imageListView) addObserver(o imageObserver) {
	v.observers = append(v.observers, o)
	o.update(v.current())
}

type imageObserver interface {
	update(i *image)
}

type repository struct {
	name          string
	uri           string
	arn           string
	tagMutability string
	createdAt     time.Time
}

func newRepository(r *ecr.Repository) *repository {
	return &repository{
		name:          aws.StringValue(r.RepositoryName),
		uri:           aws.StringValue(r.RepositoryUri),
		arn:           aws.StringValue(r.RepositoryArn),
		tagMutability: aws.StringValue(r.ImageTagMutability),
		createdAt:     aws.TimeValue(r.CreatedAt),
	}
}

func (r *repository) createdAtStr() string {
	// TODO: consider timezone
	return r.createdAt.Format(datetimeFormat)
}

type repositoryListView struct {
	cursor       int
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
	b := v.box.Enclose("REPOSITORY LIST")
	for i, r := range v.repositories {
		if v.cursor == i {
			b.Print("> ")
		} else {
			b.Print("  ")
		}
		b.Puts(r.name)
	}
}

func (v *repositoryListView) selectNext() {
	if v.cursor < len(v.repositories)-1 {
		v.cursor++
		v.notify()
	}
}

func (v *repositoryListView) selectPrev() {
	if v.cursor > 0 {
		v.cursor--
		v.notify()
	}
}

func (v *repositoryListView) selectFirst() {
	if v.cursor > 0 {
		v.cursor = 0
		v.notify()
	}
}

func (v *repositoryListView) selectLast() {
	if v.cursor < len(v.repositories)-1 {
		v.cursor = len(v.repositories) - 1
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
	return v.repositories[v.cursor]
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
