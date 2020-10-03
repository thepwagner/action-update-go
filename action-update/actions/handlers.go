package actions

import (
	"github.com/go-git/go-git/v5"
	"github.com/thepwagner/action-update/cmd"
	gitrepo "github.com/thepwagner/action-update/repo"
	"github.com/thepwagner/action-update/updater"
)

func NewHandlers(cfg *cmd.Config, f updater.Factory) cmd.HandlersByEventName {
	h := &handler{cfg: cfg, updaterFactory: f}
	return cmd.HandlersByEventName{
		"issue_comment":     IssueComment,
		"pull_request":      h.PullRequest,
		"schedule":          h.UpdateAll,
		"workflow_dispatch": h.UpdateAll,
	}
}

type handler struct {
	updaterFactory updater.Factory
	cfg            *cmd.Config
}

func (h *handler) repo() (updater.Repo, error) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return nil, err
	}
	gitRepo, err := gitrepo.NewGitRepo(repo)
	if err != nil {
		return nil, err
	}
	gitRepo.NoPush = h.cfg.NoPush

	if h.cfg.GitHubRepository == "" || h.cfg.GitHubToken == "" {
		return gitRepo, nil
	}

	return gitrepo.NewGitHubRepo(gitRepo, h.cfg.InputSigningKey, h.cfg.GitHubRepository, h.cfg.GitHubToken)
}

func (h *handler) repoUpdater(repo updater.Repo) *updater.RepoUpdater {
	return updater.NewRepoUpdater(repo, h.updaterFactory(repo.Root()))
}
