package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"
	"github.com/thepwagner/action-update-go/gomod"
	"golang.org/x/oauth2"
)

// GitHubRepo wraps GitRepo to create a GitHub PR for the pushed branch.
type GitHubRepo struct {
	repo gomod.Repo

	content  PullRequestContentFiller
	github   *github.Client
	owner    string
	repoName string
}

var _ gomod.Repo = (*GitHubRepo)(nil)

type PullRequestContentFiller func(gomod.Update) (title, body string, err error)

func NewGitHubRepo(repo gomod.Repo, repoNameOwner, token string) (*GitHubRepo, error) {
	ghRepoSplit := strings.Split(repoNameOwner, "/")
	if len(ghRepoSplit) != 2 {
		return nil, fmt.Errorf("expected repo in OWNER/NAME format")
	}

	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)
	return &GitHubRepo{
		repo:     repo,
		owner:    ghRepoSplit[0],
		repoName: ghRepoSplit[1],
		github:   github.NewClient(tc),
		content:  DefaultPullRequestContentFiller,
	}, nil
}

func (g GitHubRepo) Root() string                  { return g.repo.Root() }
func (g GitHubRepo) SetBranch(branch string) error { return g.repo.SetBranch(branch) }
func (g GitHubRepo) NewBranch(baseBranch, branch string) error {
	return g.repo.NewBranch(baseBranch, branch)
}
func (g GitHubRepo) Updates(ctx context.Context) (gomod.UpdatesByBranch, error) {
	return g.repo.Updates(ctx)
}

func (g *GitHubRepo) Push(ctx context.Context, update gomod.Update) error {
	if err := g.repo.Push(ctx, update); err != nil {
		return err
	}
	if err := g.createPR(ctx, update); err != nil {
		return err
	}
	return nil
}

func (g *GitHubRepo) createPR(ctx context.Context, update gomod.Update) error {
	title, body, err := g.content(update)
	res, _, err := g.github.PullRequests.Create(ctx, g.owner, g.repoName, &github.NewPullRequest{
		Title: &title,
		Body:  &body,
		// TODO: how to reference these neatly...
		//Base:  github.String(update.Base.Name().Short()),
		//Head:  github.String(update.BranchName()),
	})
	if err != nil {
		if strings.Contains(err.Error(), "pull request already exists") {
			return nil
		}
		return fmt.Errorf("creating PR: %w", err)
	}
	logrus.WithField("pr_id", res.ID).Info("created pull request")
	return nil
}
