package dockerurl_test

import (
	"testing"

	"github.com/thepwagner/action-update-go/updater"
	"github.com/thepwagner/action-update-go/updatertest"
)

func TestUpdater_Dependencies(t *testing.T) {
	cases := map[string][]updater.Dependency{
		"simple": {
			{Path: "github.com/containerd/containerd/releases", Version: "v1.4.0"},
		},
		"hash": {
			{Path: "github.com/elixir-lang/elixir/releases", Version: "v1.10.3"},
		},
	}

	updatertest.DependenciesFixtures(t, updaterFactory(), cases)
}
