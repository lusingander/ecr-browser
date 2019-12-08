package layout

import (
	"strings"

	"github.com/eihigh/goban"
	"github.com/lusingander/ecr-browser/util"
)

const (
	breadcrumbSep = " > "
)

type Breadcrumb struct {
	x, y, w  int
	elements []string
}

func NewBreadcrumb(x, y, w int) *Breadcrumb {
	return &Breadcrumb{x, y, w, make([]string, 0)}
}

func (b *Breadcrumb) Push(e string) {
	b.elements = append(b.elements, e)
}

func (b *Breadcrumb) Pop() string {
	e := b.elements[len(b.elements)-1]
	b.elements = util.Delete(len(b.elements)-1, b.elements)
	return e
}

func (b *Breadcrumb) View() {
	goban.NewBox(b.x, b.y, b.w, 1).Puts(strings.Join(b.elements, breadcrumbSep))
}
