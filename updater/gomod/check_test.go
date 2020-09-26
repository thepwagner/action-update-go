package gomod_test

import (
	"context"
	"testing"

	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thepwagner/action-update-go/updater"
	"github.com/thepwagner/action-update-go/updater/gomod"
)

var goGitHub29 = updater.Dependency{
	Path:    "github.com/google/go-github/v29",
	Version: "v29.0.0",
}

func TestUpdater_Check_MajorVersions(t *testing.T) {
	u := checkInFixture(t, goGitHub29, gomod.WithMajorVersions(true))
	require.NotNil(t, u)
	t.Log(u.Next)
	assert.True(t, semver.Compare("v29", u.Next) < 0)
	assert.NotEqual(t, "v29", semver.Major(u.Next))
}

func TestUpdater_Check_NotMajorVersions(t *testing.T) {
	u := checkInFixture(t, goGitHub29, gomod.WithMajorVersions(false))
	require.NotNil(t, u)
	t.Log(u.Next)
	assert.True(t, semver.Compare("v29", u.Next) < 0)
	assert.Equal(t, "v29", semver.Major(u.Next))
}

func TestUpdater_Check_MajorVersionsNotAvailable(t *testing.T) {
	t.Skip("expects v32 to be the latest, check https://github.com/google/go-github/tags before running")
	latestGoGitHubMajor := updater.Dependency{
		Path:    "github.com/google/go-github/v32",
		Version: "v32.0.0",
	}

	u := checkInFixture(t, latestGoGitHubMajor, gomod.WithMajorVersions(true))
	require.NotNil(t, u)
	t.Log(u.Next)
	assert.True(t, semver.Compare("v29", u.Next) < 0)
	assert.NotEqual(t, "v29", semver.Major(u.Next))
}

func checkInFixture(t *testing.T, dep updater.Dependency, opts ...gomod.UpdaterOpt) *updater.Update {
	u, err := updaterFromFixture(t, "simple", opts...).Check(context.Background(), dep)
	require.NoError(t, err)
	return u
}
