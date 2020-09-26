package main

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/thepwagner/action-update-go/actions"
	"github.com/thepwagner/action-update-go/cmd"
)

type HandlersByEventName map[string]actions.Handler

// Run invokes a handler by event name. Assumes Actions environment.
func Run(ctx context.Context, handlers HandlersByEventName) error {
	env, err := cmd.ParseEnvironment()
	if err != nil {
		return err
	}
	logrus.SetLevel(env.LogLevel())

	log := logrus.WithField("event_name", env.GitHubEventName)
	h, ok := handlers[env.GitHubEventName]
	if !ok {
		log.Info("unhandled event")
		return nil
	}

	evt, err := env.ParseEvent()
	if err != nil {
		return err
	}

	// Set GOPRIVATE for private modules:
	_ = os.Setenv("GOPRIVATE", "*")

	if err := h(ctx, env, evt); err != nil {
		return err
	}
	log.Debug("handler complete")
	return nil
}
