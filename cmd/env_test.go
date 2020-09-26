package cmd

import (
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
		e := Environment{InputBatches: tc.input}
		b, err := e.Batches()
		require.NoError(t, err)
		assert.Equal(t, tc.batches, b)
	}
}
