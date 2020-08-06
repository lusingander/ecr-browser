package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecr"
	"github.com/lusingander/ecr-browser/domain"
)

type awsEcrClinet struct {
	cli *ecr.ECR
}

func NewAwsEcrClient() domain.ContainerClient {
	return &awsEcrClinet{
		cli: createClient(),
	}
}

func (c *awsEcrClinet) FetchAllRepositories() ([]*domain.Repository, error) {
	return fetchRepositories(c.cli)
}

func (c *awsEcrClinet) FetchAllImages(repo string) ([]*domain.Image, error) {
	return fetchImages(c.cli, repo)
}

func newRepository(r *ecr.Repository) *domain.Repository {
	return domain.NewRepository(
		aws.StringValue(r.RepositoryName),
		aws.StringValue(r.RepositoryUri),
		aws.StringValue(r.RepositoryArn),
		aws.StringValue(r.ImageTagMutability),
		aws.TimeValue(r.CreatedAt),
	)
}

func newImage(i *ecr.ImageDetail) *domain.Image {
	return domain.NewImage(
		aws.StringValueSlice(i.ImageTags),
		aws.TimeValue(i.ImagePushedAt),
		aws.StringValue(i.ImageDigest),
		aws.Int64Value(i.ImageSizeInBytes),
	)
}

func createClient() *ecr.ECR {
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String(domain.TargetRegion),
	}))
	svc := ecr.New(sess)
	return svc
}

func fetchRepositories(svc *ecr.ECR) ([]*domain.Repository, error) {
	input := &ecr.DescribeRepositoriesInput{
		MaxResults: aws.Int64(100),
	}
	var ret []*domain.Repository
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

func fetchImages(svc *ecr.ECR, repositoryName string) ([]*domain.Image, error) {
	input := &ecr.DescribeImagesInput{
		MaxResults:     aws.Int64(100),
		RepositoryName: aws.String(repositoryName),
	}
	var ret []*domain.Image
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
