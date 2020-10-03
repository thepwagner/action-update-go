package cmd

import (
	"context"
	"fmt"

	"github.com/caarlos0/env/v6"
	"github.com/sirupsen/logrus"
)

func Execute(ctx context.Context, config interface{}, handlers HandlersByEventName) error {
	// Parse configuration:
	c, ok := config.(cfg)
	if !ok {
		return fmt.Errorf("config must embed Config")
	}
	if err := env.Parse(config); err != nil {
		return fmt.Errorf("parsing environment: %w", err)
	}
	cfg := c.cfg()

	return HandleEvent(ctx, cfg, handlers)
}

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
