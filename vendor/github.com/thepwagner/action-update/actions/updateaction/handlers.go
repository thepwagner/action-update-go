package updateaction

import (
	"github.com/go-git/go-git/v5"
	"github.com/thepwagner/action-update/actions"
	gitrepo "github.com/thepwagner/action-update/repo"
	"github.com/thepwagner/action-update/updater"
)

type HandlerParams interface {
	UpdateEnvironment
	updater.Factory
}

// NewHandlers returns Actions handlers for processing updates
func NewHandlers(p HandlerParams) *actions.Handlers {
	h := &handler{
		cfg:            p.env(),
		updaterFactory: p,
		branchNamer:    updater.DefaultUpdateBranchNamer{},
	}
	return &actions.Handlers{
		IssueComment:       IssueComment,
		PullRequest:        h.PullRequest,
		Schedule:           h.UpdateAll,
		RepositoryDispatch: h.RepositoryDispatch,
		Release:            h.Release,
		WorkflowDispatch:   h.UpdateAll,
	}
}

type handler struct {
	updaterFactory updater.Factory
	cfg            *Environment
	branchNamer    updater.UpdateBranchNamer
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

	return gitrepo.NewGitHubRepo(gitRepo, h.cfg.SigningKey(), h.cfg.GitHubRepository, h.cfg.GitHubToken)
}

func (h *handler) repoUpdater(repo updater.Repo) (*updater.RepoUpdater, error) {
	groups, err := updater.ParseGroups(h.cfg.InputGroups)
	if err != nil {
		return nil, err
	}
	return updater.NewRepoUpdater(repo, h.updaterFactory.NewUpdater(repo.Root()), updater.WithGroups(groups...), updater.WithBranchNamer(h.branchNamer)), nil
}
