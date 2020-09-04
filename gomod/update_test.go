package gomod_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thepwagner/action-update-go/gomod"
)

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
				u := gomod.Update{
					Path:     updatePath,
					Previous: baseVersion,
					Next:     v,
				}
				assert.True(t, u.MajorPkg(), v)
			}

			for _, v := range c.major {
				u := gomod.Update{
					Path:     "github.com/foo/bar",
					Previous: baseVersion,
					Next:     v,
				}
				assert.False(t, u.MajorPkg(), v)
			}
			for _, v := range c.notMajor {
				u := gomod.Update{
					Path:     updatePath,
					Previous: baseVersion,
					Next:     v,
				}
				assert.False(t, u.MajorPkg(), v)
			}
		})
	}
}
