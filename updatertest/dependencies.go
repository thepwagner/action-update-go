package updatertest

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thepwagner/action-update-go/updater"
)

type updaterFactory func(root string) updater.Updater

// DependenciesFixtures verifies .Dependencies() on an Updater initialized from a fixture.
func DependenciesFixtures(t *testing.T, factory updaterFactory, cases map[string][]updater.Dependency) {
	for fixture, expected := range cases {
		t.Run(fixture, func(t *testing.T) {
			tempDir := TempDirFromFixture(t, fixture)
			u := factory(tempDir)
			deps, err := u.Dependencies(context.Background())
			require.NoError(t, err)
			assert.Equal(t, expected, deps)
		})
	}
}
