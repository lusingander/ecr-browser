package main

import (
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/dustin/go-humanize"
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
