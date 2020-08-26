package handler

import (
	"context"

	"github.com/thepwagner/action-update-go/cmd"
)

type Handler func(context.Context, cmd.Environment, interface{}) error
