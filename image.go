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

func (i *image) display() string {
	return i.getTag()
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

type imageListView struct {
	*listViewBase
	repository string
}

func newImageListView(b *goban.Box, svc *ecr.ECR, repoName string) (*imageListView, error) {
	imgs, err := fetchImages(svc, repoName)
	if err != nil {
		return nil, err
	}
	sortImages(imgs)
	return &imageListView{
		listViewBase: &listViewBase{
			box:      b,
			elements: listViewElementsFromImages(imgs),
			title:    "IMAGES",
		},
		repository: repoName,
	}, nil
}

func listViewElementsFromImages(imgs []*image) []listViewElement {
	var elems []listViewElement
	for _, img := range imgs {
		elems = append(elems, img)
	}
	return elems
}

type imageDetailView struct {
	box      *goban.Box
	selected *image
}

func newImageDetailView(b *goban.Box) *imageDetailView {
	return &imageDetailView{b, nil}
}

func (v *imageDetailView) update(e listViewElement) {
	if img, ok := e.(*image); ok {
		v.selected = img
	}
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
