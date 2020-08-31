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

func NewGitHubRepo(repo *GitRepo, repoNameOwner, token string) (*GitHubRepo, error) {
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

func (g *GitHubRepo) Root() string                                { return g.repo.Root() }
func (g *GitHubRepo) Branch() string                              { return g.repo.Branch() }
func (g *GitHubRepo) SetBranch(branch string) error               { return g.repo.SetBranch(branch) }
func (g *GitHubRepo) Parse(branch string) (string, *gomod.Update) { return g.repo.Parse(branch) }
func (g *GitHubRepo) NewBranch(baseBranch string, update gomod.Update) error {
	return g.repo.NewBranch(baseBranch, update)
}

func (g *GitHubRepo) Updates(ctx context.Context) (gomod.UpdatesByBranch, error) {
	updates, err := g.repo.Updates(ctx)
	if err != nil {
		return nil, err
	}

	if err := g.addUpdatesFromPR(ctx, updates); err != nil {
		return nil, err
	}

	return updates, nil
}

func (g *GitHubRepo) addUpdatesFromPR(ctx context.Context, updates gomod.UpdatesByBranch) error {
	prList, _, err := g.github.PullRequests.List(ctx, g.owner, g.repoName, &github.PullRequestListOptions{
		State: "all",
	})
	if err != nil {
		return fmt.Errorf("listing pull requests: %w", err)
	}

	for _, pr := range prList {
		base, update := g.repo.Parse(*pr.Head.Ref)
		if update == nil {
			continue
		}

		logrus.WithFields(logrus.Fields{
			"base_branch": base,
			"path":        update.Path,
			"version":     update.Next,
		}).Debug("found existing PR")

		switch pr.GetState() {
		case "open":
			updates.AddOpen(base, *update)
		case "closed":
			if !pr.GetMerged() {
				updates.AddClosed(base, *update)
			}
		}
	}
	return nil
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
	if err != nil {
		return fmt.Errorf("generating PR content: %w", err)
	}

	branch := g.repo.Branch()
	baseBranch, _ := g.repo.Parse(branch)
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
