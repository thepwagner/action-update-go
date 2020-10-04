package main

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/thepwagner/action-update-go/gomodules"
	"github.com/thepwagner/action-update/actions/updateaction"
	"github.com/thepwagner/action-update/updater"
)

type Environment struct {
	updateaction.Environment
	Tidy bool `env:"INPUT_TIDY" envDefault:"true"`
}

func (c *Environment) factory(root string) updater.Updater {
	return gomodules.NewUpdater(root,
		gomodules.WithTidy(c.Tidy),
	)
}

func main() {
	// Set GOPRIVATE for private modules:
	_ = os.Setenv("GOPRIVATE", "*")

	var env Environment
	handlers := updateaction.NewHandlers(&env.Environment, env.factory)
	ctx := context.Background()
	if err := handlers.ParseAndHandle(ctx, &env); err != nil {
		logrus.WithError(err).Fatal("failed")
	}
}
