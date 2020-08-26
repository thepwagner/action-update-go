package main

import (
	"context"

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

	h, ok := handlers[env.GitHubEventName]
	if !ok {
		logrus.WithField("event_name", env.GitHubEventName).Info("unhandled event")
		return nil
	}

	evt, err := env.ParseEvent()
	if err != nil {
		return err
	}

	return h(ctx, env, evt)
}
