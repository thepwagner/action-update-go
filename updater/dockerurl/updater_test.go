package dockerurl_test

import (
	"fmt"
	"testing"

	deepcopy "github.com/otiai10/copy"
	"github.com/stretchr/testify/require"
	"github.com/thepwagner/action-update-go/updater"
	"github.com/thepwagner/action-update-go/updater/dockerurl"
)

//go:generate mockery --outpkg dockerurl_test --output . --testonly --name repoClient --structname mockRepoClient --filename mockrepoclient_test.go

const (
	previousVersion = "v1.4.0"
	nextVersion     = "v1.4.1"

	depOwner    = "containerd"
	depRepoName = "containerd"
)

var (
	depPath = fmt.Sprintf("github.com/%s/%s/releases", depOwner, depRepoName)
	dep     = updater.Dependency{Path: depPath, Version: previousVersion}
	update  = updater.Update{Path: depPath, Previous: previousVersion, Next: nextVersion}
)

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
