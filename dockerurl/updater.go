package dockerurl

import (
	"context"

	"github.com/google/go-github/v32/github"
	"github.com/thepwagner/action-update-go/updater"
)

type Updater struct {
	root   string
	github repoClient
}

type repoClient interface {
	ListReleases(ctx context.Context, owner, repo string, opts *github.ListOptions) ([]*github.RepositoryRelease, *github.Response, error)
}

var _ updater.Updater = (*Updater)(nil)

func NewUpdater(root string, gh *github.Client) *Updater {
	return &Updater{root: root, github: gh.Repositories}
}

func (u *Updater) ApplyUpdate(ctx context.Context, update updater.Update) error {
	panic("implement me")
}
