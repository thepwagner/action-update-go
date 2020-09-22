package dockerurl_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-github/v32/github"
	deepcopy "github.com/otiai10/copy"
	"github.com/stretchr/testify/require"
	"github.com/thepwagner/action-update-go/dockerurl"
	"github.com/thepwagner/action-update-go/updater"
)

func updaterFromFixture(t *testing.T, fixture string) updater.Updater {
	tempDir := tempDirFromFixture(t, fixture)
	gh := github.NewClient(http.DefaultClient)
	return dockerurl.NewUpdater(tempDir, gh)
}

func tempDirFromFixture(t *testing.T, fixture string) string {
	tempDir := t.TempDir()
	err := deepcopy.Copy(fmt.Sprintf("testdata/%s", fixture), tempDir)
	require.NoError(t, err)
	return tempDir
}
