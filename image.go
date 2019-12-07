package main

import (
	"sort"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/dustin/go-humanize"
	"github.com/eihigh/goban"
)

type image struct {
	tags     []string
	pushedAt time.Time
	digest   string
	sizeByte int64
}

func newImage(i *ecr.ImageDetail) *image {
	return &image{
		tags:     aws.StringValueSlice(i.ImageTags),
		pushedAt: aws.TimeValue(i.ImagePushedAt),
		digest:   aws.StringValue(i.ImageDigest),
		sizeByte: aws.Int64Value(i.ImageSizeInBytes),
	}
}

func (i *image) getTag() string {
	return strings.Join(i.getTags(), ", ")
}

func (i *image) getTags() []string {
	if len(i.tags) == 0 {
		return []string{"<untagged>"}
	}
	return i.tags
}

func (i *image) pushedAtStr() string {
	// TODO: consider timezone
	return i.pushedAt.Format(datetimeFormat)
}

func (i *image) sizeStr() string {
	return humanize.Bytes(uint64(i.sizeByte))
}

func imageSorter(imgs []*image) func(int, int) bool {
	return func(i, j int) bool { return imgs[i].pushedAt.After(imgs[j].pushedAt) }
}

func sortImages(imgs []*image) {
	sort.Slice(imgs, imageSorter(imgs))
}

type imageObserver interface {
	update(i *image)
}

type imageListView struct {
	*listViewBase
	images     []*image
	observers  []imageObserver
	repository string
}

func newImageListView(b *goban.Box, svc *ecr.ECR, repoName string) (*imageListView, error) {
	imgs, err := fetchImages(svc, repoName)
	if err != nil {
		return nil, err
	}
	sortImages(imgs)
	return &imageListView{
		listViewBase: &listViewBase{box: b},
		images:       imgs,
		repository:   repoName,
	}, nil
}

func (v *imageListView) View() {
	b := v.box.Enclose("IMAGE LIST")
	for i := 0; i < v.height(); i++ {
		if v.cur == i {
			b.Print("> ")
		} else {
			b.Print("  ")
		}
		if img, ok := v.get(i + v.viewTop); ok {
			b.Puts(img.getTag())
		} else {
			break
		}
	}
	createFooter(v.box, v).Print(currentCountStr(v))
}

func (v *imageListView) get(i int) (*image, bool) {
	if i >= len(v.images) {
		return nil, false
	}
	return v.images[i], true
}

func (v *imageListView) height() int {
	h := v.box.Size.Y - 2
	if len(v.images) < h {
		return len(v.images)
	}
	return h
}

func (v *imageListView) empty() bool {
	return v.images == nil || len(v.images) == 0
}

func (v *imageListView) length() int {
	return len(v.images)
}

func (v *imageListView) cursor() int {
	if v.empty() {
		return -1
	}
	return v.cur + v.viewTop
}

func (v *imageListView) cursorExistFirst() bool {
	return v.cursor() == 0
}

func (v *imageListView) cursorExistLast() bool {
	return v.cursor() == len(v.images)-1
}

func (v *imageListView) selectNext() {
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

func (v *imageListView) selectPrev() {
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

func (v *imageListView) selectFirst() {
	if v.empty() {
		return
	}
	v.cur = 0
	v.viewTop = 0
	v.notify()
}

func (v *imageListView) selectLast() {
	if v.empty() {
		return
	}
	v.cur = v.height() - 1
	v.viewTop = len(v.images) - v.height()
	v.notify()
}

func (v *imageListView) notify() {
	current := v.current()
	for _, o := range v.observers {
		o.update(current)
	}
}

func (v *imageListView) current() *image {
	if v.empty() {
		return nil
	}
	return v.images[v.cursor()]
}

func (v *imageListView) addObserver(o imageObserver) {
	v.observers = append(v.observers, o)
	o.update(v.current())
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
	}
}
