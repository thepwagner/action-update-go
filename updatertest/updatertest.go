package updatertest

import (
	"context"
	"fmt"
	"testing"

	deepcopy "github.com/otiai10/copy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thepwagner/action-update-go/updater"
)

func TempDirFromFixture(t *testing.T, fixture string) string {
	tempDir := t.TempDir()
	err := deepcopy.Copy(fmt.Sprintf("testdata/%s", fixture), tempDir)
	require.NoError(t, err)
	return tempDir
}

// Factory provides UpdaterS for testing, masking any arguments other than the repo root.
type Factory func(root string) updater.Updater

// DependenciesFixtures verifies .Dependencies() on an Updater initialized from a fixture.
func DependenciesFixtures(t *testing.T, factory Factory, cases map[string][]updater.Dependency) {
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

func CheckInFixture(t *testing.T, fixture string, factory Factory, dep updater.Dependency) *updater.Update {
	tempDir := TempDirFromFixture(t, fixture)
	u := factory(tempDir)
	update, err := u.Check(context.Background(), dep)
	require.NoError(t, err)
	return update
}

func ApplyUpdateToFixture(t *testing.T, fixture string, factory Factory, up updater.Update) string {
	tempDir := TempDirFromFixture(t, fixture)
	u := factory(tempDir)
	err := u.ApplyUpdate(context.Background(), up)
	require.NoError(t, err)
	return tempDir
}
