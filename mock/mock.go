package mock

import (
	"crypto/sha256"
	"fmt"
	"strconv"
	"time"

	"github.com/lusingander/ecr-browser/domain"
)

type mockClinet struct {
	repositoryCache
	imageCacheMap
}

type repositoryCache []*domain.Repository

type imageCacheMap map[string][]*domain.Image

func NewMockClient() domain.ContainerClient {
	return &mockClinet{
		repositoryCache: make(repositoryCache, 0),
		imageCacheMap:   make(imageCacheMap),
	}
}

func (c *mockClinet) FetchAllRepositories() ([]*domain.Repository, error) {
	if len(c.repositoryCache) > 0 {
		return c.repositoryCache, nil
	}

	time.Sleep(time.Millisecond * 500)

	n := 20
	repos := make([]*domain.Repository, 0, n)
	for i := 1; i <= n; i++ {
		repos = append(repos, repo(i))
	}

	c.repositoryCache = repos

	return repos, nil
}

func (c *mockClinet) FetchAllImages(repo string) ([]*domain.Image, error) {
	if cache, ok := c.imageCacheMap[repo]; ok {
		return cache, nil
	}

	time.Sleep(time.Millisecond * 500)

	n := 50
	images := make([]*domain.Image, 0, n)
	for i := 1; i <= n; i++ {
		images = append(images, image(i, repo))
	}

	c.imageCacheMap[repo] = images

	return images, nil
}

func repo(i int) *domain.Repository {
	name := fmt.Sprintf("sample-repo-%02d", i)
	uri := fmt.Sprintf("xxx.dkr.ecr.ap-northeast-1.amazonaws.com/%s", name)
	arn := fmt.Sprintf("arn:aws:ecr:ap-northeast-1:xxx:repository/%s", name)
	tagMutability := "MUTABLE"
	createdAt := time.Now().AddDate(0, 0, i)
	return domain.NewRepository(name, uri, arn, tagMutability, createdAt)
}

func image(i int, repo string) *domain.Image {
	is := strconv.Itoa(i)
	commit := fmt.Sprintf("%x", sha256.Sum256([]byte(repo+is)))[:7]
	var tags []string
	if i == 1 {
		tags = []string{commit, "latest"}
	} else if i%10 == 0 {
		tags = []string{}
	} else {
		tags = []string{commit}
	}
	pushedAt := time.Now().AddDate(0, 0, -i)
	digest := fmt.Sprintf("sha256:%x", sha256.Sum256([]byte(repo+is+repo)))
	sizeByte := 1024 * 1024 * i / 2
	return domain.NewImage(tags, pushedAt, digest, int64(sizeByte))
}
