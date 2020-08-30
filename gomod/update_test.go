package gomod_test

import (
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
			for _, v := range c.major {
				u := gomod.Update{Previous: baseVersion, Next: v}
				assert.True(t, u.Major(), v)
			}
			for _, v := range c.notMajor {
				u := gomod.Update{Previous: baseVersion, Next: v}
				assert.False(t, u.Major(), v)
			}
		})
	}
}
