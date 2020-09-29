package common

import "github.com/sirupsen/logrus"

func Logger() logrus.FieldLogger {
	return logrus.WithField("common", true)
}

