package main

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecr"
)

type repository struct {
	name          string
	uri           string
	arn           string
	tagMutability string
	createdAt     time.Time
}

func newRepository(r *ecr.Repository) *repository {
	return &repository{
		name:          aws.StringValue(r.RepositoryName),
		uri:           aws.StringValue(r.RepositoryUri),
		arn:           aws.StringValue(r.RepositoryArn),
		tagMutability: aws.StringValue(r.ImageTagMutability),
		createdAt:     aws.TimeValue(r.CreatedAt),
	}
}

func (r *repository) createdAtStr() string {
	// TODO: consider timezone
	return r.createdAt.Format(datetimeFormat)
}
