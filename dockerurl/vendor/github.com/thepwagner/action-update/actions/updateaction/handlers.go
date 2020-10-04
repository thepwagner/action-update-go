package updateaction

import (
	"github.com/go-git/go-git/v5"
	"github.com/thepwagner/action-update/actions"
	gitrepo "github.com/thepwagner/action-update/repo"
	"github.com/thepwagner/action-update/updater"
)

// NewHandlers returns Actions handlers for processing updates
func NewHandlers(cfg *Environment, f updater.Factory) *actions.Handlers {
	h := &handler{cfg: cfg, updaterFactory: f}
	return &actions.Handlers{
		IssueComment:     IssueComment,
		PullRequest:      h.PullRequest,
		Schedule:         h.UpdateAll,
		WorkflowDispatch: h.UpdateAll,
	}
}

type handler struct {
	updaterFactory updater.Factory
	cfg            *Environment
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
