package main

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/thepwagner/action-update-dockerurl/dockerurl"
	"github.com/thepwagner/action-update/actions/updateaction"
)

func main() {
	// Set GOPRIVATE for private modules:
	_ = os.Setenv("GOPRIVATE", "*")

	var cfg dockerurl.Environment
	handlers := updateaction.NewHandlers(&cfg)
	ctx := context.Background()
	if err := handlers.ParseAndHandle(ctx, &cfg); err != nil {
		logrus.WithError(err).Fatal("failed")
	}
}
