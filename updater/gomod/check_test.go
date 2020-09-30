package gomod_test

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thepwagner/action-update-go/updater"
	"github.com/thepwagner/action-update-go/updater/gomod"
	"github.com/thepwagner/action-update-go/updatertest"
)

var goGitHub29 = updater.Dependency{
	Path:    "github.com/google/go-github/v29",
	Version: "v29.0.0",
}

func TestUpdater_Check_MajorVersions(t *testing.T) {
	u := updatertest.CheckInFixture(t, "simple", updaterFactory(gomod.WithMajorVersions(true)), goGitHub29)
	require.NotNil(t, u)
	t.Log(u.Next)
	assert.True(t, semver.Compare("v29", u.Next) < 0)
	assert.NotEqual(t, "v29", semver.Major(u.Next))
}

func TestUpdater_Check_NotMajorVersions(t *testing.T) {
	u := updatertest.CheckInFixture(t, "simple", updaterFactory(gomod.WithMajorVersions(false)), goGitHub29)
	require.NotNil(t, u)
	t.Log(u.Next)
	assert.True(t, semver.Compare("v29", u.Next) < 0)
	assert.Equal(t, "v29", semver.Major(u.Next))
}

func TestUpdater_Check_GopkgIn(t *testing.T) {
	u := updatertest.CheckInFixture(t, "gopkg", updaterFactory(gomod.WithMajorVersions(true)), updater.Dependency{
		Path:    "gopkg.in/yaml.v1",
		Version: "v1.0.0",
	})
	require.NotNil(t, u)
	t.Log(u.Next)
	assert.True(t, semver.Compare("v1", u.Next) < 0)
	assert.NotEqual(t, "v1", semver.Major(u.Next))
}

func TestUpdater_Check_MajorVersionsNotAvailable(t *testing.T) {
	t.Skip("expects v32 to be the latest, check https://github.com/google/go-github/tags before running")
	latestGoGitHubMajor := updater.Dependency{
		Path:    "github.com/google/go-github/v32",
		Version: "v32.0.0",
	}

	u := updatertest.CheckInFixture(t, "simple", updaterFactory(gomod.WithMajorVersions(true)), latestGoGitHubMajor)
	require.NotNil(t, u)
	t.Log(u.Next)
	assert.True(t, semver.Compare("v32", u.Next) < 0)
	assert.Equal(t, "v32", semver.Major(u.Next))
}

func TestUpdater_Check_Multimodule(t *testing.T) {
	for _, path := range []string{"multimodule", filepath.Join("multimodule", "common")} {
		t.Run(strings.ReplaceAll(path, string(filepath.Separator), "-"), func(t *testing.T) {
			u := updatertest.CheckInFixture(t, path, updaterFactory(gomod.WithMajorVersions(false)), updater.Dependency{
				Path:    "github.com/pkg/errors",
				Version: "v0.8.0",
			})
			require.NotNil(t, u)
			t.Log(u.Next)
		})
	}
}
