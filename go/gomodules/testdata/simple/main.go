package main

import (
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

func main() {
	err := errors.New("kaboom")
	logrus.WithError(err).Info("")
}
