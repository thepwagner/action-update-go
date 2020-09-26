package dockerurl_test

import (
	"testing"

	"github.com/thepwagner/action-update-go/updater"
	"github.com/thepwagner/action-update-go/updater/dockerurl"
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

	updaterFactory := func(root string) updater.Updater { return dockerurl.NewUpdater(root) }
	updatertest.DependenciesFixtures(t, updaterFactory, cases)
}
