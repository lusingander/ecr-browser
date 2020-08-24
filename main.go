package main

import (
	"github.com/lusingander/ecr-browser/aws"
	"github.com/lusingander/ecr-browser/domain"
	"github.com/lusingander/ecr-browser/ui"
)

func newClient() domain.ContainerClient {
	return aws.NewAwsEcrClient()
}

func main() {
	cli := newClient()
	ui.Start(cli)
}
