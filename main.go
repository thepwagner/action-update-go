package main

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/thepwagner/action-update-go/gomodules"
	"github.com/thepwagner/action-update/actions/updateaction"
)

func main() {
	// Set GOPRIVATE for private modules:
	_ = os.Setenv("GOPRIVATE", "*")

	ctx := context.Background()

	var env gomodules.Environment
	handlers := updateaction.NewHandlers(&env)
	if err := handlers.ParseAndHandle(ctx, &env); err != nil {
		logrus.WithError(err).Fatal("failed")
	}
}
