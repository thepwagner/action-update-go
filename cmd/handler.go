package cmd

import (
	"context"
)

// Handler responds to a deserialized GitHub event
type Handler func(ctx context.Context, env *Environment, evt interface{}) error
