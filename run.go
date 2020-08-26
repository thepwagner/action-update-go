package main

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
	"github.com/thepwagner/action-update-go/cmd"
	"github.com/thepwagner/action-update-go/handler"
)

type HandlersByEventName map[string]handler.Handler

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
