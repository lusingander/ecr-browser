package domain

import (
	"sort"
	"time"
)

type Repository struct {
	Name          string
	Uri           string
	Arn           string
	TagMutability string
	CreatedAt     time.Time
}

func NewRepository(name string, uri string, arn string, tagMutability string, createdAt time.Time) *Repository {
	return &Repository{
		Name:          name,
		Uri:           uri,
		Arn:           arn,
		TagMutability: tagMutability,
		CreatedAt:     createdAt,
	}
}

func (r *Repository) Display() string {
	return r.Name
}

func (r *Repository) CreatedAtStr() string {
	// TODO: consider timezone
	return r.CreatedAt.Format(datetimeFormat)
}

func repositorySorter(repos []*Repository) func(int, int) bool {
	return func(i, j int) bool { return repos[i].Name < repos[j].Name }
}

func SortRepositories(repos []*Repository) {
	sort.Slice(repos, repositorySorter(repos))
}
