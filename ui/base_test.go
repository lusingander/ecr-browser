package ui

import (
	"testing"

	"github.com/eihigh/goban"
)

func TestViewStack_push(t *testing.T) {
	sut := newViewStack()

	vs1 := []goban.View{
		&DummyView{id: 1},
	}
	sut.push(vs1...)
	vs2 := []goban.View{
		&DummyView{id: 2},
		&DummyView{id: 3},
	}
	sut.push(vs2...)

	got := sut.length()
	if got != 2 {
		t.Errorf("length() = %v; want = %v", got, 2)
	}
}

func TestViewStack_pop(t *testing.T) {
	sut := newViewStack()

	vs1 := []goban.View{
		&DummyView{id: 1},
	}
	sut.push(vs1...)
	vs2 := []goban.View{
		&DummyView{id: 2},
		&DummyView{id: 3},
	}
	sut.push(vs2...)

	got := sut.pop()
	if len(got) != 2 {
		t.Errorf("len(%v) = %v; want = %v", got, len(got), 2)
	}
	if v, ok := got[0].(*DummyView); !ok || v.id != 2 {
		t.Errorf("%v, %v = pop()[0]; want %v, %v", v, ok, true, vs2[0])
	}
	if v, ok := got[1].(*DummyView); !ok || v.id != 3 {
		t.Errorf("%v, %v = pop()[1]; want %v, %v", v, ok, true, vs2[1])
	}
}

type DummyView struct {
	id int
}

func (v *DummyView) View() {}
