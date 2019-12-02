package main

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/endpoints"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/eihigh/goban"
	"github.com/gdamore/tcell"
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

func createClient() *ecr.ECR {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(targetRegion),
	}))
	svc := ecr.New(sess)
	return svc
}

func app(_ context.Context, es goban.Events) error {
	svc := createClient()

	mainView := newMainView()
	listView, err := newRepositoryListView(mainView.grid.list, svc)
	if err != nil {
		return err
	}
	detailView := newRepositoryDetailView(mainView.grid.detail)
	listView.addObserver(detailView)
	goban.PushView(mainView)
	goban.PushView(listView)
	goban.PushView(detailView)

	for {
		goban.Show()

		k := es.ReadKey()
		switch k.Rune() {
		case 'k':
			listView.selectPrev()
		case 'j':
			listView.selectNext()
		case 'g':
			listView.selectFirst()
		case 'G':
			listView.selectLast()
		default:
			switch k.Key() {
			case tcell.KeyUp:
				listView.selectPrev()
			case tcell.KeyDown:
				listView.selectNext()
			default:
				return nil
			}
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
	return &repositoryListView{
		box:          b,
		repositories: repos,
	}, nil
}

func fetchRepositories(svc *ecr.ECR) ([]*repository, error) {
	input := &ecr.DescribeRepositoriesInput{
		MaxResults: aws.Int64(100),
	}
	output, err := svc.DescribeRepositories(input)
	if err != nil {
		return nil, err
	}
	var ret []*repository
	// TODO: consider NextToken
	for _, r := range output.Repositories {
		ret = append(ret, newRepository(r))
	}
	return ret, nil
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
