package gomod_test

import (
	"context"
	"testing"

	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/modfile"
	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/module"
	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/semver"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thepwagner/action-update-go/gomod"
)

var goGitHub29 = &modfile.Require{
	Mod: module.Version{
		Path:    "github.com/google/go-github/v29",
		Version: "v29.0.0",
	},
}

func TestUpdateChecker_CheckForModuleUpdates_Major(t *testing.T) {
	checker := &gomod.UpdateChecker{
		MajorVersions: true,
		RootDir:       "../fixtures/simple",
	}

	ctx := context.Background()
	update, err := checker.CheckForModuleUpdates(ctx, goGitHub29)
	require.NoError(t, err)
	require.NotNil(t, update)
	t.Log(update.Next)
	assert.True(t, semver.Compare("v29", update.Next) < 0)
	assert.NotEqual(t, "v29", semver.Major(update.Next))
}

func TestUpdateChecker_CheckForModuleUpdates_MajorNotAvailable(t *testing.T) {
	t.Skip("expects to never release")
	checker := &gomod.UpdateChecker{
		MajorVersions: true,
		RootDir:       "../fixtures/simple",
	}

	ctx := context.Background()
	update, err := checker.CheckForModuleUpdates(ctx, &modfile.Require{
		Mod: module.Version{
			Path:    "github.com/google/go-github/v32",
			Version: "v32.0.0",
		},
	})
	require.NoError(t, err)
	require.NotNil(t, update)
	t.Log(update.Next)
	assert.True(t, semver.Compare("v29", update.Next) < 0)
	assert.NotEqual(t, "v29", semver.Major(update.Next))
}

func TestUpdateChecker_CheckForModuleUpdates(t *testing.T) {
	checker := &gomod.UpdateChecker{
		MajorVersions: false,
		RootDir:       "../fixtures/simple",
	}

	ctx := context.Background()
	update, err := checker.CheckForModuleUpdates(ctx, goGitHub29)
	require.NoError(t, err)
	require.NotNil(t, update)
	t.Log(update.Next)
	assert.True(t, semver.Compare("v29", update.Next) < 0)
	assert.Equal(t, "v29", semver.Major(update.Next))
}
