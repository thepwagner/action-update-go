package cmd

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/thepwagner/action-update-go/handler"
)

type HandlersByEventName map[string]handler.Handler

func Run(ctx context.Context, handlers HandlersByEventName) error {
	env, err := ParseEnvironment()
	if err != nil {
		return err
	}

	h, ok := handlers[env.GitHubEventName]
	if !ok {
		logrus.WithField("event_name", env.GitHubEventName).Info("unhandled event")
		return nil
	}

	evt, err := env.parseEvent()
	if err != nil {
		return err
	}

	return h(ctx, evt)
}
