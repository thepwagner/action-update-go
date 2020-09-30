package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"
	"github.com/thepwagner/action-update-go/updater"
	"golang.org/x/oauth2"
)

// GitHubRepo wraps GitRepo to create a GitHub PR for the pushed branch.
type GitHubRepo struct {
	repo updater.Repo

	prContent PullRequestContent
	github    *github.Client
	owner     string
	repoName  string
}

var _ updater.Repo = (*GitHubRepo)(nil)

type PullRequestContent interface {
	Generate(context.Context, ...updater.Update) (title, body string, err error)
}

func NewGitHubRepo(repo *GitRepo, repoNameOwner, token string) (*GitHubRepo, error) {
	ghRepoSplit := strings.Split(repoNameOwner, "/")
	if len(ghRepoSplit) != 2 {
		return nil, fmt.Errorf("expected repo in OWNER/NAME format")
	}

	ghClient := NewGitHubClient(token)
	return &GitHubRepo{
		repo:      repo,
		owner:     ghRepoSplit[0],
		repoName:  ghRepoSplit[1],
		github:    ghClient,
		prContent: NewGitHubPullRequestContent(ghClient),
	}, nil
}

func NewGitHubClient(token string) *github.Client {
	ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
	tc := oauth2.NewClient(context.Background(), ts)
	ghClient := github.NewClient(tc)
	return ghClient
}

func (g *GitHubRepo) Root() string                        { return g.repo.Root() }
func (g *GitHubRepo) Branch() string                      { return g.repo.Branch() }
func (g *GitHubRepo) SetBranch(branch string) error       { return g.repo.SetBranch(branch) }
func (g *GitHubRepo) NewBranch(base, branch string) error { return g.repo.NewBranch(base, branch) }

// Push follows the git push with opening a pull request
func (g *GitHubRepo) Push(ctx context.Context, updates ...updater.Update) error {
	if err := g.repo.Push(ctx, updates...); err != nil {
		return err
	}
	if err := g.createPR(ctx, updates); err != nil {
		return err
	}
	return nil
}

func (g *GitHubRepo) createPR(ctx context.Context, updates []updater.Update) error {
	title, body, err := g.prContent.Generate(ctx, updates...)
	if err != nil {
		return fmt.Errorf("generating PR prContent: %w", err)
	}

	branch := g.repo.Branch()
	baseBranch := strings.Split(branch, "/")[1]
	pullRequest, _, err := g.github.PullRequests.Create(ctx, g.owner, g.repoName, &github.NewPullRequest{
		Title: &title,
		Body:  &body,
		Base:  &baseBranch,
		Head:  &branch,
	})
	if err != nil {
		if strings.Contains(err.Error(), "pull request already exists") {
			return nil
		}
		return fmt.Errorf("creating PR: %w", err)
	}
	logrus.WithField("pr_id", *pullRequest.ID).Info("created pull request")
	return nil
}
