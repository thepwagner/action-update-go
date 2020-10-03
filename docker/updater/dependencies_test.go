package updater_test

import (
	"testing"

	"github.com/thepwagner/action-update-go/updatertest"
	updater2 "github.com/thepwagner/action-update/updater"
)

func TestUpdater_Dependencies(t *testing.T) {
	cases := map[string][]updater2.Dependency{
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
