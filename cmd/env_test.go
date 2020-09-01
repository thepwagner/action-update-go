package cmd

import (
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestEnvironment_LogLevel(t *testing.T) {
	cases := map[string]logrus.Level{
		"":        logrus.InfoLevel,
		"invalid": logrus.InfoLevel,
		"warn":    logrus.WarnLevel,
	}

	for in, lvl := range cases {
		t.Run(fmt.Sprintf("parse %q", in), func(t *testing.T) {
			e := Environment{InputLogLevel: in}
			assert.Equal(t, lvl, e.LogLevel())
		})
	}
}
