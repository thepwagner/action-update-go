package handler

import (
	"context"

	"github.com/go-git/go-git/v5"
	"github.com/thepwagner/action-update-go/cmd"
	"github.com/thepwagner/action-update-go/gomod"
)

func Schedule(_ context.Context, env cmd.Environment, _ interface{}) error {
	repo, err := git.PlainOpen(".")
	if err != nil {
		return err
	}
	updater, err := gomod.NewUpdater(repo)
	if err != nil {
		return err
	}

	for _, b := range env.Branches() {
		if err := updater.UpdateAll(b); err != nil {
			return err
		}
	}
	return nil
}

var _ Handler = Schedule
