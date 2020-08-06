package domain

import (
	"sort"
	"strings"
	"time"

	"github.com/dustin/go-humanize"
)

type Image struct {
	Tags     []string
	PushedAt time.Time
	Digest   string
	SizeByte int64
}

func NewImage(tags []string, pushedAt time.Time, digest string, sizeByte int64) *Image {
	return &Image{
		Tags:     tags,
		PushedAt: pushedAt,
		Digest:   digest,
		SizeByte: sizeByte,
	}
}

func (i *Image) Display() string {
	return i.GetTag()
}

func (i *Image) GetTag() string {
	return strings.Join(i.GetTags(), ", ")
}

func (i *Image) GetTags() []string {
	if len(i.Tags) == 0 {
		return []string{noTag}
	}
	return i.Tags
}

func (i *Image) PushedAtStr() string {
	// TODO: consider timezone
	return i.PushedAt.Format(datetimeFormat)
}

func (i *Image) SizeStr() string {
	return humanize.Bytes(uint64(i.SizeByte))
}

func imageSorter(imgs []*Image) func(int, int) bool {
	return func(i, j int) bool { return imgs[i].PushedAt.After(imgs[j].PushedAt) }
}

func SortImages(imgs []*Image) {
	sort.Slice(imgs, imageSorter(imgs))
}
