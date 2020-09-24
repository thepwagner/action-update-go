package dockerurl

import (
	"context"
	"net/http"

	"github.com/google/go-github/v32/github"
	"github.com/thepwagner/action-update-go/updater"
)

type Updater struct {
	root    string
	ghRepos repoClient
}

type repoClient interface {
	ListReleases(ctx context.Context, owner, repo string, opts *github.ListOptions) ([]*github.RepositoryRelease, *github.Response, error)
}

var _ updater.Updater = (*Updater)(nil)

type UpdaterOpt func(*Updater)

// WithRepoClient provides a go-github RepositoriesService or suitable mock.
func WithRepoClient(rc repoClient) UpdaterOpt {
	return func(u *Updater) {
		u.ghRepos = rc
	}
}

func NewUpdater(root string, opts ...UpdaterOpt) *Updater {
	u := &Updater{root: root}
	for _, opt := range opts {
		opt(u)
	}

	if u.ghRepos == nil {
		gh := github.NewClient(http.DefaultClient)
		u.ghRepos = gh.Repositories
	}

	return u
}

func (u *Updater) ApplyUpdate(ctx context.Context, update updater.Update) error {
	panic("implement me")
}
