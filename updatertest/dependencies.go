package updatertest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thepwagner/action-update-go/updater"
)

type updaterFactory func(t *testing.T, fixture string) updater.Updater

func DependenciesFixtures(t *testing.T, factory updaterFactory, cases map[string][]updater.Dependency) {
	for fixture, expected := range cases {
		t.Run(fixture, func(t *testing.T) {
			u := factory(t, fixture)
			deps, err := u.Dependencies(context.Background())
			require.NoError(t, err)
			assert.Equal(t, expected, deps)
		})
	}
}
