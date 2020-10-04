package actions

import (
	"context"
	"fmt"

	"github.com/caarlos0/env/v6"
	"github.com/sirupsen/logrus"
)

// Execute loads `config` from environment, then selects the
func Execute(ctx context.Context, config interface{}, handlers HandlersByEventName) error {
	ac, ok := config.(actionsConfig)
	if !ok {
		return fmt.Errorf("config should embed actions.Config")
	}

	if err := env.Parse(config); err != nil {
		return fmt.Errorf("parsing environment: %w", err)
	}
	cfg := ac.cfg()
	return HandleEvent(ctx, cfg, handlers)
}

// HandleEvent invokes a handler by name.
func HandleEvent(ctx context.Context, cfg *Config, handlers HandlersByEventName) error {
	// Is there a handler for this event?
	log := logrus.WithField("event_name", cfg.GitHubEventName)
	h, ok := handlers[cfg.GitHubEventName]
	if !ok {
		log.Info("unhandled event")
		return nil
	}

	// Parse event and invoke handler:
	evt, err := cfg.ParseEvent()
	if err != nil {
		return err
	}
	if err := h(ctx, evt); err != nil {
		return err
	}
	log.Debug("handler complete")
	return nil
}
