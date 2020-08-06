package domain

type ContainerClient interface {
	FetchAllRepositories() ([]*Repository, error)
	FetchAllImages(repo string) ([]*Image, error)
}
