package main

import (
	"fmt"
	"strconv"
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

type cursorer interface {
	length() int
	cursor() int // return -1 if length is zero
}

func createFooter(b *goban.Box, c cursorer) *goban.Box {
	l := calcCountStrMaxLen(c)
	h := 1
	y := b.Pos.Y + b.Size.Y - h
	x := b.Pos.X + b.Size.X - l - 1 // right justify
	w := l
	return goban.NewBox(x, y, w, h)
}

func calcCountStrMaxLen(c cursorer) int {
	return len(countStr(c, c.length()))
}

func currentCountStr(c cursorer) string {
	return countStr(c, c.cursor()+1)
}

func countStr(c cursorer, n int) string {
	l := c.length()
	d := len(strconv.Itoa(l))
	return fmt.Sprintf(countFormat, d, n, d, l)
}

type listView interface {
	goban.View
	selectNext()
	selectPrev()
	selectFirst()
	selectLast()
}

type listViewBase struct {
	cur       int
	box       *goban.Box
	viewTop   int
	elements  []listViewElement
	title     string
	observers []listElementObserver
}

func (v *listViewBase) View() {
	b := v.box.Enclose(v.title)
	for i := 0; i < v.height(); i++ {
		if v.cur == i {
			b.Print("> ")
		} else {
			b.Print("  ")
		}
		if e, ok := v.get(i + v.viewTop); ok {
			b.Puts(e.display())
		} else {
			break
		}
	}
	createFooter(v.box, v).Print(currentCountStr(v))
}

func (v *listViewBase) get(i int) (listViewElement, bool) {
	if i >= len(v.elements) {
		return nil, false
	}
	return v.elements[i], true
}

func (v *listViewBase) height() int {
	h := v.box.Size.Y - 2
	if len(v.elements) < h {
		return len(v.elements)
	}
	return h
}

func (v *listViewBase) empty() bool {
	return v.elements == nil || len(v.elements) == 0
}

func (v *listViewBase) length() int {
	return len(v.elements)
}

func (v *listViewBase) current() listViewElement {
	if v.empty() {
		return nil
	}
	return v.elements[v.cursor()]
}

func (v *listViewBase) cursor() int {
	if v.empty() {
		return -1
	}
	return v.cur + v.viewTop
}

func (v *listViewBase) cursorExistFirst() bool {
	return v.cursor() == 0
}

func (v *listViewBase) cursorExistLast() bool {
	return v.cursor() == len(v.elements)-1
}

func (v *listViewBase) selectNext() {
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

func (v *listViewBase) selectPrev() {
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

func (v *listViewBase) selectFirst() {
	if v.empty() {
		return
	}
	v.cur = 0
	v.viewTop = 0
	v.notify()
}

func (v *listViewBase) selectLast() {
	if v.empty() {
		return
	}
	v.cur = v.height() - 1
	v.viewTop = len(v.elements) - v.height()
	v.notify()
}

func (v *listViewBase) addObserver(o listElementObserver) {
	v.observers = append(v.observers, o)
	o.update(v.current())
}

func (v *listViewBase) notify() {
	current := v.current()
	for _, o := range v.observers {
		o.update(current)
	}
}

type listViewElement interface {
	display() string
}

type listElementObserver interface {
	update(e listViewElement)
}

type detailView interface {
	goban.View
}
