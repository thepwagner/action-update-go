package main

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/thepwagner/action-update-dockerurl/dockerurl"
	"github.com/thepwagner/action-update/actions/updateaction"
	"github.com/thepwagner/action-update/updater"
)

type Environment struct {
	updateaction.Environment
}

func (c *Environment) factory(root string) updater.Updater {
	return dockerurl.NewUpdater(root)
}

func main() {
	// Set GOPRIVATE for private modules:
	_ = os.Setenv("GOPRIVATE", "*")

	var cfg Environment
	handlers := updateaction.NewHandlers(&cfg.Environment, cfg.factory)
	ctx := context.Background()
	if err := handlers.ParseAndHandle(ctx, &cfg); err != nil {
		logrus.WithError(err).Fatal("failed")
	}
}
