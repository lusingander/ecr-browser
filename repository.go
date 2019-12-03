package main

import (
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

type repositoryObserver interface {
	update(r *repository)
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
