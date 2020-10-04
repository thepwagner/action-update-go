package main

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/thepwagner/action-update-dockerurl/dockerurl"
	"github.com/thepwagner/action-update/actions"
	"github.com/thepwagner/action-update/actions/update"
	"github.com/thepwagner/action-update/updater"
)

type Config struct {
	update.Config
	Tidy bool `env:"INPUT_TIDY" envDefault:"true"`
}

func (c *Config) factory(root string) updater.Updater {
	return dockerurl.NewUpdater(root)
}

func main() {
	// Set GOPRIVATE for private modules:
	_ = os.Setenv("GOPRIVATE", "*")

	var cfg Config
	handlers := update.NewHandlers(&cfg.Config, cfg.factory)
	ctx := context.Background()
	if err := actions.Execute(ctx, &cfg, handlers); err != nil {
		logrus.WithError(err).Fatal("failed")
	}
}
