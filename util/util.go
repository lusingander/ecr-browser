package util

import "github.com/eihigh/goban"

func PushViews(vs ...goban.View) {
	for _, v := range vs {
		goban.PushView(v)
	}
}

func RemoveViews(vs ...goban.View) {
	for _, v := range vs {
		goban.RemoveView(v)
	}
}

func InsideSides(src *goban.Box, l, t, r, b int) *goban.Box {
	// InsideSides does not word as expected :sob:
	if l == 0 && t == 0 && r == 0 && b == 0 {
		return src
	}
	tmp := src.InsideSides(l, t, r, b)
	if l > 0 {
		l--
	}
	if t > 0 {
		t--
	}
	if r > 0 {
		r--
	}
	if b > 0 {
		b--
	}
	return InsideSides(tmp, l, t, r, b)
}

func Delete(i int, s []string) []string {
	return append(s[:i], s[i+1:]...)
}
