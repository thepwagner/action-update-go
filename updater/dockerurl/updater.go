package dockerurl

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/go-github/v32/github"
	"github.com/thepwagner/action-update-go/updater"
)

type Updater struct {
	root    string
	http    *http.Client
	ghRepos repoClient
}

// repoClient is a sub-interface of go-github RepositoryService
type repoClient interface {
	ListReleases(ctx context.Context, owner, repo string, opts *github.ListOptions) ([]*github.RepositoryRelease, *github.Response, error)
	GetReleaseByTag(ctx context.Context, owner, repo, tag string) (*github.RepositoryRelease, *github.Response, error)
}

var _ updater.Updater = (*Updater)(nil)

type UpdaterOpt func(*Updater)

// WithRepoClient provides a go-github RepositoriesService or suitable mock.
func WithRepoClient(rc repoClient) UpdaterOpt {
	return func(u *Updater) {
		u.ghRepos = rc
	}
}

// NewUpdater creates a new Updater
func NewUpdater(root string, opts ...UpdaterOpt) *Updater {
	u := &Updater{
		root: root,
		http: http.DefaultClient,
	}
	for _, opt := range opts {
		opt(u)
	}

	if u.ghRepos == nil {
		gh := github.NewClient(u.http)
		u.ghRepos = gh.Repositories
	}

	return u
}

func formatGitHubRelease(repo, name string) string {
	return fmt.Sprintf("github.com/%s/%s/releases", repo, name)
}

func parseGitHubRelease(path string) (owner, repoNme string) {
	pathSplit := strings.SplitN(path, "/", 4)
	return pathSplit[1], pathSplit[2]
}
