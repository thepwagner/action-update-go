package dockerurl_test

import (
	"testing"

	"github.com/thepwagner/action-update-go/updater"
	"github.com/thepwagner/action-update-go/updatertest"
)

func TestUpdater_Dependencies(t *testing.T) {
	cases := map[string][]updater.Dependency{
		"simple": {
			{Path: "https://github.com/containerd/containerd/releases", Version: "v1.4.0"},
		},
	}
	updatertest.DependenciesFixtures(t, updaterFromFixture, cases)
}
