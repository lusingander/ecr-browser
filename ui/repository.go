package ui

import (
	"fmt"

	"github.com/eihigh/goban"
	"github.com/gdamore/tcell"
	"github.com/lusingander/ecr-browser/domain"
	"github.com/pkg/browser"
)

const (
	repositoryListViewTitle = "REPOSITORIES"
)

type repositoryListView struct {
	*listViewBase
}

func newRepositoryListView(b *goban.Box, cli domain.ContainerClient) (*repositoryListView, error) {
	repos, err := cli.FetchAllRepositories()
	if err != nil {
		return nil, err
	}
	domain.SortRepositories(repos)
	return &repositoryListView{
		listViewBase: &listViewBase{
			box:      b,
			elements: listViewElementsFromRepositories(repos),
			title:    repositoryListViewTitle,
		},
	}, nil
}

func listViewElementsFromRepositories(repos []*domain.Repository) []listViewElement {
	var elems []listViewElement
	for _, repo := range repos {
		elems = append(elems, repo)
	}
	return elems
}

func (v *repositoryListView) operate(key *tcell.EventKey) {
	switch key.Rune() {
	case 'l':
		v.parent.displayImageViews(v.currentRepositoryName())
	case 'o':
		v.openWebBrowser()
	default:
		v.listViewBase.operate(key)
	}
}

func (v *repositoryListView) currentRepositoryName() string {
	if repo, ok := v.current().(*domain.Repository); ok {
		return repo.Name
	}
	return ""
}

func (v *repositoryListView) openWebBrowser() error {
	repo := v.currentRepositoryName()
	url := createECRConsoleRepositoryURL(repo)
	return browser.OpenURL(url)
}

type repositoryDetailView struct {
	box      *goban.Box
	selected *domain.Repository
}

func newRepositoryDetailView(b *goban.Box) *repositoryDetailView {
	return &repositoryDetailView{b, nil}
}

func (v *repositoryDetailView) update(e listViewElement) {
	if repo, ok := e.(*domain.Repository); ok {
		v.selected = repo
	}
}

func (v *repositoryDetailView) View() {
	b := v.box.Enclose("DETAIL")
	if v.selected != nil {
		b.Puts("NAME:")
		b.Puts("  " + v.selected.Name)
		b.Puts("URI:")
		b.Puts("  " + v.selected.Uri)
		b.Puts("ARN:")
		b.Puts("  " + v.selected.Arn)
		b.Puts("TAG MUTABILITY:")
		b.Puts("  " + v.selected.TagMutability)
		b.Puts("CREATED AT:")
		b.Puts("  " + v.selected.CreatedAtStr())
	}
}

func createECRConsoleURL() string {
	region := domain.TargetRegion
	url := "https://%s.console.aws.amazon.com/ecr/repositories?region=%s"
	return fmt.Sprintf(url, region, region)
}

func createECRConsoleRepositoryURL(repo string) string {
	region := domain.TargetRegion
	url := "https://%s.console.aws.amazon.com/ecr/repositories/%s/?region=%s"
	return fmt.Sprintf(url, region, repo, region)
}
