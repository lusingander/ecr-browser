package main

import "github.com/eihigh/goban"

type detailView interface {
	goban.View
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

type imageDetailView struct {
	box      *goban.Box
	selected *image
}

func newImageDetailView(b *goban.Box) *imageDetailView {
	return &imageDetailView{b, nil}
}

func (v *imageDetailView) update(i *image) {
	v.selected = i
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
		// TODO: show current repository name
	}
}
