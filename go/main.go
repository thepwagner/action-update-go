package main

import (
	"context"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/thepwagner/action-update-go/gomodules"
	"github.com/thepwagner/action-update/actions"
	"github.com/thepwagner/action-update/cmd"
	"github.com/thepwagner/action-update/updater"
)

type Config struct {
	cmd.Config
	Tidy bool `env:"INPUT_TIDY" envDefault:"true"`
}

func (c *Config) factory(root string) updater.Updater {
	fmt.Println(c.GitHubEventName, c.Tidy)
	return gomodules.NewUpdater(root)
}

func main() {
	// Set GOPRIVATE for private modules:
	_ = os.Setenv("GOPRIVATE", "*")

	var cfg Config
	ctx := context.Background()
	handlers := actions.NewHandlers(&cfg.Config, cfg.factory)
	if err := cmd.Execute(ctx, &cfg, handlers); err != nil {
		logrus.WithError(err).Fatal("failed")
	}
}
