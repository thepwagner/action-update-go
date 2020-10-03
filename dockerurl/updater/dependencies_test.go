package updater_test

import (
	"testing"

	"github.com/thepwagner/action-update-go/updatertest"
	updater2 "github.com/thepwagner/action-update/updater"
)

func TestUpdater_Dependencies(t *testing.T) {
	cases := map[string][]updater2.Dependency{
		"simple": {
			{Path: "github.com/containerd/containerd/releases", Version: "v1.4.0"},
		},
		"hash": {
			{Path: "github.com/elixir-lang/elixir/releases", Version: "v1.10.3"},
		},
	}

	updatertest.DependenciesFixtures(t, updaterFactory(), cases)
}
