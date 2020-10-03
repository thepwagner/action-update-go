package cmd_test

import (
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thepwagner/action-update/cmd"
)

func TestConfig_LogLevel(t *testing.T) {
	cases := map[string]logrus.Level{
		"":        logrus.InfoLevel,
		"invalid": logrus.InfoLevel,
		"warn":    logrus.WarnLevel,
	}

	for in, lvl := range cases {
		t.Run(fmt.Sprintf("parse %q", in), func(t *testing.T) {
			e := cmd.Config{InputLogLevel: in}
			assert.Equal(t, lvl, e.LogLevel())
		})
	}
}

func TestConfig_Batches(t *testing.T) {
	cases := []struct {
		input   string
		batches map[string][]string
	}{
		{
			input:   `foo: [bar, baz]`,
			batches: map[string][]string{"foo": {"bar", "baz"}},
		},
		{
			input: `---
foo: bar
foz: baz`,
			batches: map[string][]string{
				"foo": {"bar"},
				"foz": {"baz"},
			},
		},
		{
			input: `foo:
- bar
- baz`,
			batches: map[string][]string{
				"foo": {"bar", "baz"},
			},
		},
		{
			input:   "",
			batches: map[string][]string{},
		},
	}

	for _, tc := range cases {
		e := cmd.Config{InputBatches: tc.input}
		b, err := e.Batches()
		require.NoError(t, err)
		assert.Equal(t, tc.batches, b)
	}
}
