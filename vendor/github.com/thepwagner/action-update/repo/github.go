package repo

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/google/go-github/v36/github"
	"github.com/sirupsen/logrus"
	"github.com/thepwagner/action-update/updater"
	"golang.org/x/oauth2"
)

// GitHubRepo wraps GitRepo to create a GitHub PR for the pushed branch.
type GitHubRepo struct {
	repo    *GitRepo
	hmacKey []byte

	prContent PullRequestContent
	github    *github.Client
	owner     string
	repoName  string
}

var _ updater.Repo = (*GitHubRepo)(nil)

type PullRequestContent interface {
	Generate(context.Context, updater.UpdateGroup) (title, body string, err error)
}

func NewGitHubRepo(repo *GitRepo, hmacKey []byte, repoNameOwner, token string) (*GitHubRepo, error) {
	ghRepoSplit := strings.Split(repoNameOwner, "/")
	if len(ghRepoSplit) != 2 {
		return nil, fmt.Errorf("expected repo in OWNER/NAME format")
	}

	ghClient := NewGitHubClient(token)
	return &GitHubRepo{
		repo:      repo,
		hmacKey:   hmacKey,
		owner:     ghRepoSplit[0],
		repoName:  ghRepoSplit[1],
		github:    ghClient,
		prContent: NewGitHubPullRequestContent(ghClient, hmacKey),
	}, nil
}

func NewGitHubClient(token string) *github.Client {
	var client *http.Client
	if token != "" {
		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		client = oauth2.NewClient(context.Background(), ts)
	} else {
		client = http.DefaultClient
	}
	return github.NewClient(client)
}

func (g *GitHubRepo) Root() string                        { return g.repo.Root() }
func (g *GitHubRepo) Branch() string                      { return g.repo.Branch() }
func (g *GitHubRepo) SetBranch(branch string) error       { return g.repo.SetBranch(branch) }
func (g *GitHubRepo) NewBranch(base, branch string) error { return g.repo.NewBranch(base, branch) }

func (g *GitHubRepo) Fetch(ctx context.Context, branch string) error {
	return g.repo.Fetch(ctx, branch)
}

// Push follows the git push with opening a pull request
func (g *GitHubRepo) Push(ctx context.Context, updates updater.UpdateGroup) error {
	if err := g.repo.Push(ctx, updates); err != nil {
		return err
	}
	if g.repo.NoPush {
		return nil
	}

	if err := g.createPR(ctx, updates); err != nil {
		return err
	}
	return nil
}

func (g *GitHubRepo) createPR(ctx context.Context, updates updater.UpdateGroup) error {
	title, body, err := g.prContent.Generate(ctx, updates)
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
	g.repoLogger().WithField("pr_number", pullRequest.GetNumber()).Info("created pull request")
	return nil
}

func (g *GitHubRepo) ExistingUpdates(ctx context.Context, baseBranch string) (updater.ExistingUpdates, error) {
	return g.updateFromPR(ctx, baseBranch)
}

func (g *GitHubRepo) updateFromPR(ctx context.Context, baseBranch string) (updater.ExistingUpdates, error) {
	// List pull requests targeting this branch:
	prs, _, err := g.github.PullRequests.List(ctx, g.owner, g.repoName, &github.PullRequestListOptions{
		State: "all",
		Base:  baseBranch,
	})
	if err != nil {
		return nil, fmt.Errorf("listing pull requests: %w", err)
	}
	g.repoLogger().WithField("pr_count", len(prs)).Debug("listed pull requests")

	var updates []updater.ExistingUpdate
	for _, pr := range prs {
		// Filter PullRequestS with signed messages:
		prLog := logrus.WithField("pr_number", pr.GetNumber())
		updateGroup, err := g.verifiedUpdateGroup(pr)
		if err != nil {
			prLog.WithError(err).Warn("signature failed")
			continue
		} else if updateGroup == nil {
			prLog.Debug("not an update PR")
			continue
		}

		// Combine signed payload with PR state to form an ExistingUpdate:
		existing := updater.ExistingUpdate{
			BaseBranch: pr.GetBase().GetRef(),
			LastUpdate: pr.GetUpdatedAt(),
			Group:      *updateGroup,
		}
		switch pr.GetState() {
		case "open":
			existing.Open = true
		case "closed":
			existing.Merged = !pr.GetMergedAt().IsZero()
		}
		if logrus.IsLevelEnabled(logrus.DebugLevel) {
			u := make([]string, 0, len(existing.Group.Updates))
			for _, update := range existing.Group.Updates {
				u = append(u, update.Path)
			}
			prLog.WithFields(logrus.Fields{
				"open":    existing.Open,
				"merged":  existing.Merged,
				"group":   existing.Group.Name,
				"updates": u,
			}).Debug("existing update added")
		}
		updates = append(updates, existing)
	}
	return updates, nil
}

func (g *GitHubRepo) verifiedUpdateGroup(pr *github.PullRequest) (*updater.UpdateGroup, error) {
	signed := ExtractSignedUpdateGroup(pr.GetBody())
	if signed == nil {
		return nil, nil
	}
	return updater.VerifySignedUpdateGroup(g.hmacKey, *signed)
}

func (g *GitHubRepo) repoLogger() *logrus.Entry {
	return logrus.WithFields(logrus.Fields{"repo_owner": g.owner, "repo_name": g.repoName})
}
