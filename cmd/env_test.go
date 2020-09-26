package cmd

import (
	"fmt"
	"strings"
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

func TestEnvironment_Batches(t *testing.T) {
	cases := []struct {
		lines   []string
		batches map[string][]string
	}{
		{
			lines:   []string{"foo:bar,baz"},
			batches: map[string][]string{"foo": {"bar", "baz"}},
		},
		{
			lines: []string{"foo:bar", "foz:baz"},
			batches: map[string][]string{
				"foo": {"bar"},
				"foz": {"baz"},
			},
		},

		{
			lines: []string{"  foo : bar,  baz"},
			batches: map[string][]string{
				"foo": {"bar", "baz"},
			},
		},
		{
			lines:   []string{"no colon, ignored"},
			batches: map[string][]string{},
		},
	}

	for _, tc := range cases {
		e := Environment{InputBatches: strings.Join(tc.lines, "\n")}
		assert.Equal(t, tc.batches, e.Batches())
	}
}
