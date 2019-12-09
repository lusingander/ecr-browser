package main

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
)

func createClient() *ecr.ECR {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(targetRegion),
	}))
	svc := ecr.New(sess)
	return svc
}

func fetchRepositories(svc *ecr.ECR) ([]*repository, error) {
	input := &ecr.DescribeRepositoriesInput{
		MaxResults: aws.Int64(100),
	}
	var ret []*repository
	for {
		output, err := svc.DescribeRepositories(input)
		if err != nil {
			return nil, err
		}
		for _, r := range output.Repositories {
			ret = append(ret, newRepository(r))
		}
		nextToken := aws.StringValue(output.NextToken)
		if nextToken == "" {
			break
		}
		input.SetNextToken(nextToken)
	}
	return ret, nil
}

func fetchImages(svc *ecr.ECR, repositoryName string) ([]*image, error) {
	input := &ecr.DescribeImagesInput{
		MaxResults:     aws.Int64(100),
		RepositoryName: aws.String(repositoryName),
	}
	var ret []*image
	for {
		output, err := svc.DescribeImages(input)
		if err != nil {
			return nil, err
		}
		for _, i := range output.ImageDetails {
			ret = append(ret, newImage(i))
		}
		nextToken := aws.StringValue(output.NextToken)
		if nextToken == "" {
			break
		}
		input.SetNextToken(nextToken)
	}
	return ret, nil
}

func createECRConsoleURL(region string) string {
	url := "https://%s.console.aws.amazon.com/ecr/repositories?region=%s"
	return fmt.Sprintf(url, region, region)
}

func createECRConsoleRepositoryURL(region string, repo string) string {
	url := "https://%s.console.aws.amazon.com/ecr/repositories/%s/?region=%s"
	return fmt.Sprintf(url, region, repo, region)
}
