package docker_test

import (
	"fmt"
	"testing"

	deepcopy "github.com/otiai10/copy"
	"github.com/stretchr/testify/require"
	"github.com/thepwagner/action-update-go/docker"
)

func updaterFromFixture(t *testing.T, fixture string) *docker.Updater {
	tempDir := tempDirFromFixture(t, fixture)
	return docker.NewUpdater(tempDir)
}

func tempDirFromFixture(t *testing.T, fixture string) string {
	tempDir := t.TempDir()
	err := deepcopy.Copy(fmt.Sprintf("testdata/%s", fixture), tempDir)
	require.NoError(t, err)
	return tempDir
}
