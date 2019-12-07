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

func (r *repository) display() string {
	return r.name
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

type repositoryListView struct {
	*listViewBase
}

func newRepositoryListView(b *goban.Box, svc *ecr.ECR) (*repositoryListView, error) {
	repos, err := fetchRepositories(svc)
	if err != nil {
		return nil, err
	}
	sortRepositories(repos)
	return &repositoryListView{
		listViewBase: &listViewBase{
			box:      b,
			elements: listViewElementsFromRepositories(repos),
			title:    "REPOSITORIES",
		},
	}, nil
}

func listViewElementsFromRepositories(repos []*repository) []listViewElement {
	var elems []listViewElement
	for _, repo := range repos {
		elems = append(elems, repo)
	}
	return elems
}

func (v *repositoryListView) currentRepositoryName() string {
	if repo, ok := v.current().(*repository); ok {
		return repo.name
	}
	return ""
}

type repositoryDetailView struct {
	box      *goban.Box
	selected *repository
}

func newRepositoryDetailView(b *goban.Box) *repositoryDetailView {
	return &repositoryDetailView{b, nil}
}

func (v *repositoryDetailView) update(e listViewElement) {
	if repo, ok := e.(*repository); ok {
		v.selected = repo
	}
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
