package main

import (
	"github.com/lusingander/ecr-browser/aws"
	"github.com/lusingander/ecr-browser/ui"
)

func main() {
	cli := aws.NewAwsEcrClient()
	ui.Start(cli)
}
