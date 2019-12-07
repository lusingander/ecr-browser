package main

import (
	"sort"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/eihigh/goban"
)

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

func repositorySorter(repos []*repository) func(int, int) bool {
	return func(i, j int) bool { return repos[i].name < repos[j].name }
}

func sortRepositories(repos []*repository) {
	sort.Slice(repos, repositorySorter(repos))
}

type repositoryObserver interface {
	update(r *repository)
}

type repositoryListView struct {
	*listViewBase
	repositories []*repository
	observers    []repositoryObserver
}

func newRepositoryListView(b *goban.Box, svc *ecr.ECR) (*repositoryListView, error) {
	repos, err := fetchRepositories(svc)
	if err != nil {
		return nil, err
	}
	sortRepositories(repos)
	return &repositoryListView{
		listViewBase: &listViewBase{box: b},
		repositories: repos,
	}, nil
}

func (v *repositoryListView) View() {
	b := v.box.Enclose("REPOSITORY LIST")
	for i := 0; i < v.height(); i++ {
		if v.cur == i {
			b.Print("> ")
		} else {
			b.Print("  ")
		}
		if r, ok := v.get(i + v.viewTop); ok {
			b.Puts(r.name)
		} else {
			break
		}
	}
	createFooter(v.box, v).Print(currentCountStr(v))
}

func (v *repositoryListView) get(i int) (*repository, bool) {
	if i >= len(v.repositories) {
		return nil, false
	}
	return v.repositories[i], true
}

func (v *repositoryListView) height() int {
	h := v.box.Size.Y - 2
	if len(v.repositories) < h {
		return len(v.repositories)
	}
	return h
}

func (v *repositoryListView) empty() bool {
	return v.repositories == nil || len(v.repositories) == 0
}

func (v *repositoryListView) length() int {
	return len(v.repositories)
}

func (v *repositoryListView) cursor() int {
	if v.empty() {
		return -1
	}
	return v.cur + v.viewTop
}

func (v *repositoryListView) cursorExistFirst() bool {
	return v.cursor() == 0
}

func (v *repositoryListView) cursorExistLast() bool {
	return v.cursor() == len(v.repositories)-1
}

func (v *repositoryListView) selectNext() {
	if v.empty() {
		return
	}
	if v.cur < v.height()-1 {
		v.cur++
	} else {
		if !v.cursorExistLast() {
			v.viewTop++
		}
	}
	v.notify()
}

func (v *repositoryListView) selectPrev() {
	if v.empty() {
		return
	}
	if v.cur > 0 {
		v.cur--
	} else {
		if !v.cursorExistFirst() {
			v.viewTop--
		}
	}
	v.notify()
}

func (v *repositoryListView) selectFirst() {
	if v.empty() {
		return
	}
	v.cur = 0
	v.viewTop = 0
	v.notify()
}

func (v *repositoryListView) selectLast() {
	if v.empty() {
		return
	}
	v.cur = v.height() - 1
	v.viewTop = len(v.repositories) - v.height()
	v.notify()
}

func (v *repositoryListView) notify() {
	current := v.current()
	for _, o := range v.observers {
		o.update(current)
	}
}

func (v *repositoryListView) current() *repository {
	if v.empty() {
		return nil
	}
	return v.repositories[v.cursor()]
}

func (v *repositoryListView) currentRepositoryName() string {
	return v.current().name
}

func (v *repositoryListView) addObserver(o repositoryObserver) {
	v.observers = append(v.observers, o)
	o.update(v.current())
}

type repositoryDetailView struct {
	box      *goban.Box
	selected *repository
}

func newRepositoryDetailView(b *goban.Box) *repositoryDetailView {
	return &repositoryDetailView{b, nil}
}

func (v *repositoryDetailView) update(r *repository) {
	v.selected = r
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
