package main

import (
	"fmt"
	"strconv"

	"github.com/eihigh/goban"
)

type listViewBase struct {
	cur       int
	box       *goban.Box
	elements  []listViewElement
	observers []listElementObserver
	title     string
	viewTop   int
}

type listViewElement interface {
	display() string
}

type listElementObserver interface {
	update(e listViewElement)
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
	v.printScroll()
	v.createFooter().Print(v.currentCountStr())
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

func (v *listViewBase) createFooter() *goban.Box {
	b := v.box
	l := v.calcCountStrMaxLen()
	h := 1
	y := b.Pos.Y + b.Size.Y - h
	x := b.Pos.X + b.Size.X - l - 1 // right justify
	w := l
	return goban.NewBox(x, y, w, h)
}

func (v *listViewBase) calcCountStrMaxLen() int {
	return len(v.countStr(v.length()))
}

func (v *listViewBase) currentCountStr() string {
	return v.countStr(v.cursor() + 1)
}

func (v *listViewBase) countStr(n int) string {
	l := v.length()
	d := len(strconv.Itoa(l))
	return fmt.Sprintf(countFormat, d, n, d, l)
}

func (v *listViewBase) printScroll() {
	l := len(v.elements)
	if l == 0 || l < v.height() {
		return
	}
	b := v.box
	barLength := v.height() * v.height() / l
	x := b.Pos.X + b.Size.X - 2
	y := b.Pos.Y + 1 + v.calcScrollTopPos(barLength)
	w := 1
	s := goban.NewBox(x, y, w, barLength)
	for i := 0; i < barLength; i++ {
		s.Puts("│") // ║ (2551) or │ (2503)
	}
}

func (v *listViewBase) calcScrollTopPos(barLength int) int {
	barMaxMove := float64(v.height() - barLength)
	viewTopMaxMove := float64(len(v.elements) - v.height())
	currentViewTop := float64(v.viewTop)
	return int((currentViewTop / viewTopMaxMove) * barMaxMove)
}
