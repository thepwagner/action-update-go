package updateaction_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thepwagner/action-update/actions/updateaction"
	"github.com/thepwagner/action-update/updater"
)

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
		e := updateaction.Environment{InputBatches: tc.input}
		b, err := e.Batches()
		require.NoError(t, err)
		assert.Equal(t, tc.batches, b)
	}
}

type testEnvironment struct {
	updateaction.Environment
}

func (t *testEnvironment) NewUpdater(root string) updater.Updater { return nil }
