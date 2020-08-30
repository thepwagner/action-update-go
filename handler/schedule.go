package handler

import (
	"context"

	"github.com/go-git/go-git/v5"
	"github.com/sirupsen/logrus"
	"github.com/thepwagner/action-update-go/cmd"
	"github.com/thepwagner/action-update-go/gomod"
	gitrepo "github.com/thepwagner/action-update-go/repo"
)

func Schedule(ctx context.Context, env cmd.Environment, _ interface{}) error {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return err
	}
	gitRepo, err := gitrepo.NewGitRepo(repo)
	if err != nil {
		return err
	}

	var modRepo gomod.Repo
	if env.GitHubRepository != "" && env.GitHubToken != "" {
		modRepo, err = gitrepo.NewGitHubRepo(gitRepo, env.GitHubRepository, env.GitHubToken)
		if err != nil {
			return err
		}
	} else {
		modRepo = gitRepo
	}

	updater, err := gomod.NewRepoUpdater(modRepo)
	if err != nil {
		return err
	}

	initialBranch := gitRepo.Branch()
	defer func() {
		if err := gitRepo.SetBranch(initialBranch); err != nil {
			logrus.WithError(err).Warn("error reverting to initial branch")
		}
	}()

	// If branches were provided as input, target those:
	if branches := env.Branches(); len(branches) > 0 {
		if err := updater.UpdateAll(ctx, branches...); err != nil {
			return err
		}
		return nil
	}

	// No branches as input, fallback to current branch:
	if err := updater.UpdateAll(ctx, initialBranch); err != nil {
		return err
	}
	return nil
}

var _ Handler = Schedule
