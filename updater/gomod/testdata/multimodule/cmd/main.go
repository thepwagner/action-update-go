package main

import (
	"github.com/pkg/errors"
	"github.com/thepwagner/action-update-go/multimodule/common"
)

func main() {
	err := errors.New("kaboom")
	common.Logger().WithError(err).Info("")
}
