package ui

import (
	"github.com/eihigh/goban"
	"github.com/gdamore/tcell"
	"github.com/lusingander/ecr-browser/domain"
)

const (
	imageListViewTitle = "IMAGES"
)

type imageListView struct {
	*listViewBase
	repository string
}

func newImageListView(b *goban.Box, repoName string) (*imageListView, error) {
	imgs, err := client.FetchAllImages(repoName)
	if err != nil {
		return nil, err
	}
	domain.SortImages(imgs)
	return &imageListView{
		listViewBase: &listViewBase{
			box:      b,
			elements: listViewElementsFromImages(imgs),
			title:    imageListViewTitle,
		},
		repository: repoName,
	}, nil
}

func listViewElementsFromImages(imgs []*domain.Image) []listViewElement {
	var elems []listViewElement
	for _, img := range imgs {
		elems = append(elems, img)
	}
	return elems
}

func (v *imageListView) operate(key *tcell.EventKey) {
	switch key.Rune() {
	case 'h':
		v.ui.loadRepositoryView(false)
	default:
		v.listViewBase.operate(key)
	}
}

type imageDetailView struct {
	box      *goban.Box
	selected *domain.Image
}

func newImageDetailView(b *goban.Box) *imageDetailView {
	return &imageDetailView{b, nil}
}

func (v *imageDetailView) update(e listViewElement) {
	if img, ok := e.(*domain.Image); ok {
		v.selected = img
	}
}

func (v *imageDetailView) View() {
	b := v.box.Enclose("DETAIL")
	if v.selected != nil {
		b.Puts("TAGS:")
		for _, t := range v.selected.GetTags() {
			b.Puts("  " + t)
		}
		b.Puts("PUSHED AT:")
		b.Puts("  " + v.selected.PushedAtStr())
		b.Puts("DIGEST:")
		b.Puts("  " + v.selected.Digest)
		b.Puts("SIZE:")
		b.Puts("  " + v.selected.SizeStr())
	}
}
