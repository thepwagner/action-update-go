package dockerurl_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/go-github/v32/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/thepwagner/action-update-go/updater"
	"github.com/thepwagner/action-update-go/updater/dockerurl"
	"github.com/thepwagner/action-update-go/updatertest"
)

func TestUpdater_Check(t *testing.T) {
	cases := map[string]struct {
		releases []*github.RepositoryRelease
		update   *string
	}{
		"no update available": {
			releases: []*github.RepositoryRelease{
				{TagName: github.String(previousVersion)},
			},
			update: nil,
		},
		"new version available": {
			releases: []*github.RepositoryRelease{
				{TagName: github.String(nextVersion)},
			},
			update: github.String(nextVersion),
		},
		"pre-release version available": {
			releases: []*github.RepositoryRelease{
				{TagName: github.String(nextVersion), Prerelease: github.Bool(true)},
			},
			update: nil,
		},
		"many versions available": {
			releases: []*github.RepositoryRelease{
				{TagName: github.String("v1.3.1")},
				{TagName: github.String(nextVersion)},
				{TagName: github.String("v1.3.0")},
				{TagName: github.String("my-awesome-release")},
			},
			update: github.String(nextVersion),
		},
	}

	ctx := context.Background()
	for lbl, tc := range cases {
		t.Run(lbl, func(t *testing.T) {
			rc := &mockRepoClient{}
			rc.On("ListReleases", ctx, depOwner, depRepoName, mock.Anything).Return(tc.releases, nil, nil)
			u := dockerurl.NewUpdater("", dockerurl.WithRepoClient(rc))

			update, err := u.Check(ctx, dep)
			require.NoError(t, err)
			if tc.update == nil {
				assert.Nil(t, update)
			} else {
				assert.Equal(t, &updater.Update{
					Path:     depPath,
					Previous: previousVersion,
					Next:     *tc.update,
				}, update)
			}

			rc.AssertExpectations(t)
		})
	}
}

func TestUpdater_Check_Unknown(t *testing.T) {
	u := dockerurl.NewUpdater("")
	_, err := u.Check(context.Background(), updater.Dependency{Path: "foo.com/bar"})
	assert.Error(t, err)
}

func TestUpdater_Check_Error(t *testing.T) {
	ctx := context.Background()
	listErr := errors.New("kaboom")
	rc := &mockRepoClient{}
	rc.On("ListReleases", ctx, depOwner, depRepoName, mock.Anything).Return(nil, nil, listErr)
	u := dockerurl.NewUpdater("", dockerurl.WithRepoClient(rc))

	_, err := u.Check(ctx, dep)
	assert.Equal(t, listErr, errors.Unwrap(err))
	rc.AssertExpectations(t)
}

func TestUpdater_CheckLive(t *testing.T) {
	t.Skip("hardcodes assumptions about latest releases")

	u := updaterFromFixture(t, "simple")
	cases := []struct {
		dep  updater.Dependency
		next *string
	}{
		{
			dep: updater.Dependency{
				Path:    "github.com/containerd/containerd/releases",
				Version: "v1.4.0",
			},
			next: github.String("v1.4.1"),
		},
		{
			dep: updater.Dependency{
				Path:    "github.com/containerd/containerd/releases",
				Version: "v1.4.1",
			},
		},
		{
			dep: updater.Dependency{
				Path:    "github.com/torvalds/linux/releases",
				Version: "v5.8",
			},
		},
		{
			dep: updater.Dependency{
				Path:    "github.com/hashicorp/terraform/releases",
				Version: "v0.13.0",
			},
			next: github.String("v0.13.3"),
		},
		{
			dep: updater.Dependency{
				Path:    "github.com/elixir-lang/elixir/releases",
				Version: "v1.10.3",
			},
			next: github.String("v1.10.4"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.dep.Path, func(t *testing.T) {
			upd, err := u.Check(context.Background(), tc.dep)
			require.NoError(t, err)

			if tc.next == nil {
				assert.Nil(t, upd)
			} else if assert.NotNil(t, upd, "no update") {
				assert.Equal(t, *tc.next, upd.Next)
			}
		})
	}
}

func updaterFromFixture(t *testing.T, fixture string) updater.Updater {
	tempDir := updatertest.TempDirFromFixture(t, fixture)
	return dockerurl.NewUpdater(tempDir)
}
