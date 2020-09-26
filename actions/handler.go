package actions

import (
	"context"

	"github.com/thepwagner/action-update-go/cmd"
)

// Handler responds to a deserialized GitHub event
type Handler func(ctx context.Context, env *cmd.Environment, evt interface{}) error
