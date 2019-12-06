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

type imageObserver interface {
	update(i *image)
}

type imageListView struct {
	cur        int
	box        *goban.Box
	images     []*image
	observers  []imageObserver
	repository string
}

func newImageListView(b *goban.Box, svc *ecr.ECR, repoName string) (*imageListView, error) {
	imgs, err := fetchImages(svc, repoName)
	if err != nil {
		return nil, err
	}
	sort.Slice(imgs, imageSorter(imgs))
	return &imageListView{
		box:        b,
		images:     imgs,
		repository: repoName,
	}, nil
}

func (v *imageListView) View() {
	// TODO: scroll / paging
	b := v.box.Enclose("IMAGE LIST")
	for i, img := range v.images {
		if v.cur == i {
			b.Print("> ")
		} else {
			b.Print("  ")
		}
		b.Puts(img.getTag())
	}
	createFooter(v.box, v).Print(currentCountStr(v))
}

func (v *imageListView) length() int {
	return len(v.images)
}

func (v *imageListView) cursor() int {
	return v.cur
}

func (v *imageListView) selectNext() {
	if v.cur < len(v.images)-1 {
		v.cur++
		v.notify()
	}
}

func (v *imageListView) selectPrev() {
	if v.cur > 0 {
		v.cur--
		v.notify()
	}
}

func (v *imageListView) selectFirst() {
	if v.cur > 0 {
		v.cur = 0
		v.notify()
	}
}

func (v *imageListView) selectLast() {
	if v.cur < len(v.images)-1 {
		v.cur = len(v.images) - 1
		v.notify()
	}
}

func (v *imageListView) notify() {
	current := v.current()
	for _, o := range v.observers {
		o.update(current)
	}
}

func (v *imageListView) current() *image {
	return v.images[v.cur]
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
		// TODO: show current repository name
	}
}
