package layout

import "github.com/eihigh/goban"

const (
	message = "Now Loading..."
)

type LoadingDialog struct {
	parent *goban.Box
	ch     chan bool
}

func NewLoadingDialog(parent *goban.Box) *LoadingDialog {
	return &LoadingDialog{parent, make(chan bool)}
}

func (d *LoadingDialog) View() {
	dialog := goban.NewBox(0, 0, len(message)+10, 7).CenterOf(d.parent).Enclose("")
	strArea := goban.NewBox(0, 0, len(message), 1).CenterOf(dialog)
	strArea.Puts(message)
}

func (d *LoadingDialog) Display() {
	goban.PushView(d)
	defer goban.RemoveView(d)
	goban.Show()
	// should catch es(goban.Events) and process(discard?)
	<-d.ch
}

func (d *LoadingDialog) Close() {
	d.ch <- true
}

func (d *LoadingDialog) WaitFor(f func()) {
	go d.Display()
	defer d.Close()
	f()
}
