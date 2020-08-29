package handler

import (
	"context"

	"github.com/go-git/go-git/v5"
	"github.com/thepwagner/action-update-go/cmd"
	gitrepo "github.com/thepwagner/action-update-go/git"
	"github.com/thepwagner/action-update-go/gomod"
)

func Schedule(ctx context.Context, env cmd.Environment, _ interface{}) error {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return err
	}

	sharedRepo, err := gitrepo.NewSharedRepo(repo)
	if err != nil {
		return err
	}

	updater, err := gomod.NewUpdater(sharedRepo, env.GitHubRepository, env.GitHubToken)
	if err != nil {
		return err
	}

	for _, b := range env.Branches() {
		if err := updater.UpdateAll(ctx, b); err != nil {
			return err
		}
	}
	return nil
}

var _ Handler = Schedule
