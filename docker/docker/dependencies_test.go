package docker_test

import (
	"testing"

	"github.com/thepwagner/action-update/updater"
	"github.com/thepwagner/action-update/updatertest"
)

func TestUpdater_Dependencies(t *testing.T) {
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
	updatertest.DependenciesFixtures(t, updaterFactory(), cases)
}
