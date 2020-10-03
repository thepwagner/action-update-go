package cmd

import (
	"context"
)

// Handler responds to a deserialized GitHub event
type Handler func(ctx context.Context, evt interface{}) error

type HandlersByEventName map[string]Handler
