package main

import (
	"strings"

	"github.com/eihigh/goban"
)

const (
	breadcrumbSep = " > "
)

type breadcrumb struct {
	x, y, w  int
	elements []string
}

func newBreadcrumb(x, y, w int) *breadcrumb {
	return &breadcrumb{x, y, w, make([]string, 0)}
}

func (b *breadcrumb) push(e string) {
	b.elements = append(b.elements, e)
}

func (b *breadcrumb) pop() string {
	e := b.elements[len(b.elements)-1]
	b.elements = delete(len(b.elements)-1, b.elements)
	return e
}

func (b *breadcrumb) View() {
	goban.NewBox(b.x, b.y, b.w, 1).Puts(strings.Join(b.elements, breadcrumbSep))
}
