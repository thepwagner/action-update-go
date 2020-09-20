package docker_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thepwagner/action-update-go/updater"
)

func TestUpdater_Dependencies(t *testing.T) {
	// fixture name to expected dependencies
	cases := map[string][]updater.Dependency{
		"simple": {
			{Path: "alpine", Version: "3.11.0"},
		},
		"buildarg": {
			{Path: "redis", Version: "6.0.0-alpine"},
			{Path: "redis", Version: "6.0.0-alpine"},
			{Path: "alpine", Version: "3.11.0"},
		},
	}

	for fixture, expected := range cases {
		t.Run(fixture, func(t *testing.T) {
			u := updaterFromFixture(t, fixture)

			deps, err := u.Dependencies(context.Background())
			require.NoError(t, err)
			assert.Equal(t, expected, deps)
		})
	}
}
