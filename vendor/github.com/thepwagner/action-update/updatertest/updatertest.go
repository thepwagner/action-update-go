package updatertest

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"

	deepcopy "github.com/otiai10/copy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thepwagner/action-update/updater"
)

func TempDirFromFixture(t *testing.T, fixture string) string {
	tempDir := t.TempDir()
	err := deepcopy.Copy(fmt.Sprintf("testdata/%s", fixture), tempDir)
	require.NoError(t, err)
	return tempDir
}

// DependenciesFixtures verifies .Dependencies() on an Updater initialized from a fixture.
func DependenciesFixtures(t *testing.T, factory updater.Factory, cases map[string][]updater.Dependency) {
	for fixture, expected := range cases {
		t.Run(fixture, func(t *testing.T) {
			tempDir := TempDirFromFixture(t, fixture)
			u := factory.NewUpdater(tempDir)
			deps, err := u.Dependencies(context.Background())
			require.NoError(t, err)
			assert.Equal(t, expected, deps)
		})
	}
}

func CheckInFixture(t *testing.T, fixture string, factory updater.Factory, dep updater.Dependency, filter func(string) bool) *updater.Update {
	tempDir := TempDirFromFixture(t, fixture)
	u := factory.NewUpdater(tempDir)
	update, err := u.Check(context.Background(), dep, filter)
	require.NoError(t, err)
	return update
}

func ApplyUpdateToFixture(t *testing.T, fixture string, factory updater.Factory, up updater.Update) string {
	dir, f := filepath.Split(fixture)

	var tempDir string
	var u updater.Updater
	if dir != "" {
		tempDir = TempDirFromFixture(t, dir)
		u = factory.NewUpdater(filepath.Join(tempDir, f))
	} else {
		tempDir = TempDirFromFixture(t, f)
		u = factory.NewUpdater(tempDir)
	}

	err := u.ApplyUpdate(context.Background(), up)
	require.NoError(t, err)
	return tempDir
}
