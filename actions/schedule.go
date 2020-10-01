package actions

import (
	"context"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/sirupsen/logrus"
	"github.com/thepwagner/action-update-go/cmd"
	gitrepo "github.com/thepwagner/action-update-go/repo"
	"github.com/thepwagner/action-update-go/updater"
	"github.com/thepwagner/action-update-go/updater/dockerurl"
	"github.com/thepwagner/action-update-go/updater/gomod"
)

func Schedule(ctx context.Context, env *cmd.Environment, _ interface{}) error {
	repo, upd, err := getRepoUpdater(env)
	if err != nil {
		return err
	}

	initialBranch := repo.Branch()
	defer func() {
		if err := repo.SetBranch(initialBranch); err != nil {
			logrus.WithError(err).Warn("error reverting to initial branch")
		}
	}()

	// If branches were provided as input, target those:
	if branches := env.Branches(); len(branches) > 0 {
		if err := upd.UpdateAll(ctx, branches...); err != nil {
			return err
		}
		return nil
	}

	// No branches as input, fallback to current branch:
	if err := upd.UpdateAll(ctx, initialBranch); err != nil {
		return err
	}
	return nil
}

var _ cmd.Handler = Schedule

func getRepoUpdater(env *cmd.Environment) (updater.Repo, *updater.RepoUpdater, error) {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return nil, nil, err
	}
	gitRepo, err := gitrepo.NewGitRepo(repo)
	if err != nil {
		return nil, nil, err
	}

	var modRepo updater.Repo
	if env.GitHubRepository != "" && env.GitHubToken != "" {
		modRepo, err = gitrepo.NewGitHubRepo(gitRepo, env.InputSigningKey, env.GitHubRepository, env.GitHubToken)
		if err != nil {
			return nil, nil, err
		}
	} else {
		modRepo = gitRepo
	}

	modUpdater := getUpdater(modRepo.Root(), env.InputUpdater)
	batches, err := env.Batches()
	if err != nil {
		return nil, nil, fmt.Errorf("parsing batches")
	}
	repoUpdater := updater.NewRepoUpdater(modRepo, modUpdater, updater.WithBatches(batches))
	return gitRepo, repoUpdater, nil
}

func getUpdater(root, updaterName string) updater.Updater {
	switch updaterName {
	case "dockerurl":
		return dockerurl.NewUpdater(root)
	case "", "gomod", "gomodules":
		return gomod.NewUpdater(root)
	default:
		logrus.WithField("updater", updaterName).Warn("unknown updater, defaulting to go modules")
		return gomod.NewUpdater(root)
	}
}
