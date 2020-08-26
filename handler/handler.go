package handler

import "context"

type Handler func(context.Context, interface{}) error
