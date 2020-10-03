package main

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/thepwagner/action-update/actions"
	"github.com/thepwagner/action-update/cmd"
)

func main() {
	ctx := context.Background()
	if err := cmd.Run(ctx, actions.Handlers); err != nil {
		logrus.WithError(err).Fatal("failed")
	}
}
