package gomod_test

import (
	"fmt"
	"testing"

	deepcopy "github.com/otiai10/copy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thepwagner/action-update-go/gomod"
	"github.com/thepwagner/action-update-go/updater"
)

func updaterFromFixture(t *testing.T, fixture string, opts ...gomod.UpdaterOpt) *gomod.Updater {
	tempDir := tempDirFromFixture(t, fixture)
	return gomod.NewUpdater(tempDir, opts...)
}

func tempDirFromFixture(t *testing.T, fixture string) string {
	tempDir := t.TempDir()
	err := deepcopy.Copy(fmt.Sprintf("testdata/%s", fixture), tempDir)
	require.NoError(t, err)
	return tempDir
}

func TestUpdate_Major(t *testing.T) {
	cases := map[string]struct {
		major    []string
		notMajor []string
	}{
		"v1": {
			major:    []string{"v2", "v2.1.1"},
			notMajor: []string{"v1", "v1.1"},
		},
		"v2": {
			major: []string{"v1", "v3"},
		},
	}

	for baseVersion, c := range cases {
		t.Run(baseVersion, func(t *testing.T) {
			updatePath := fmt.Sprintf("github.com/foo/bar/%s", baseVersion)
			for _, v := range c.major {
				u := updater.Update{
					Path:     updatePath,
					Previous: baseVersion,
					Next:     v,
				}
				assert.True(t, gomod.MajorPkg(u), v)
			}

			for _, v := range c.major {
				u := updater.Update{
					Path:     "github.com/foo/bar",
					Previous: baseVersion,
					Next:     v,
				}
				assert.False(t, gomod.MajorPkg(u), v)
			}
			for _, v := range c.notMajor {
				u := updater.Update{
					Path:     updatePath,
					Previous: baseVersion,
					Next:     v,
				}
				assert.False(t, gomod.MajorPkg(u), v)
			}
		})
	}
}
