package main

import (
	"flag"
	"log"

	"github.com/lusingander/ecr-browser/aws"
	"github.com/lusingander/ecr-browser/domain"
	"github.com/lusingander/ecr-browser/mock"
	"github.com/lusingander/ecr-browser/ui"
)

var (
	useMock *bool
)

func parseFlags() {
	useMock = flag.Bool("mock", false, "Use mock data")
	flag.Parse()
}

func newClient() domain.ContainerClient {
	if *useMock {
		return mock.NewMockClient()
	}
	return aws.NewAwsEcrClient()
}

func main() {
	parseFlags()
	cli := newClient()
	if err := ui.Start(cli); err != nil {
		log.Fatal(err)
	}
}
