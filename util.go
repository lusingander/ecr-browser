package main

import "github.com/eihigh/goban"

func pushViews(vs ...goban.View) {
	for _, v := range vs {
		goban.PushView(v)
	}
}

func removeViews(vs ...goban.View) {
	for _, v := range vs {
		goban.RemoveView(v)
	}
}
