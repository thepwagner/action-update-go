package cmd

import (
	"context"
	"os"

	"github.com/sirupsen/logrus"
)

type HandlersByEventName map[string]Handler

// Run invokes a handler by event name. Assumes Actions environment.
func Run(ctx context.Context, handlers HandlersByEventName) error {
	env, err := ParseEnvironment()
	if err != nil {
		return err
	}
	logrus.SetLevel(env.LogLevel())
	return HandleEvent(ctx, env, handlers)
}

func HandleEvent(ctx context.Context, env *Environment, handlers HandlersByEventName) error {
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
