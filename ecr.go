package main

import (
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
	output, err := svc.DescribeRepositories(input)
	if err != nil {
		return nil, err
	}
	var ret []*repository
	// TODO: consider NextToken
	for _, r := range output.Repositories {
		ret = append(ret, newRepository(r))
	}
	return ret, nil
}

func fetchImages(svc *ecr.ECR, repositoryName string) ([]*image, error) {
	input := &ecr.DescribeImagesInput{
		MaxResults:     aws.Int64(100),
		RepositoryName: aws.String(repositoryName),
	}
	output, err := svc.DescribeImages(input)
	if err != nil {
		return nil, err
	}
	var ret []*image
	// TODO: consider NextToken
	for _, i := range output.ImageDetails {
		ret = append(ret, newImage(i))
	}
	return ret, nil
}
