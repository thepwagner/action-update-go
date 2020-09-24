package dockerurl_test

import (
	"fmt"
	"testing"

	deepcopy "github.com/otiai10/copy"
	"github.com/stretchr/testify/require"
	"github.com/thepwagner/action-update-go/dockerurl"
	"github.com/thepwagner/action-update-go/updater"
)

//go:generate mockery --outpkg dockerurl_test --output . --testonly --name repoClient --structname mockRepoClient --filename mockrepoclient_test.go

func updaterFromFixture(t *testing.T, fixture string) updater.Updater {
	tempDir := tempDirFromFixture(t, fixture)
	return dockerurl.NewUpdater(tempDir)
}

func tempDirFromFixture(t *testing.T, fixture string) string {
	tempDir := t.TempDir()
	err := deepcopy.Copy(fmt.Sprintf("testdata/%s", fixture), tempDir)
	require.NoError(t, err)
	return tempDir
}
