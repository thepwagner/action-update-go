package gomod_test

import (
	"testing"

	"github.com/thepwagner/action-update-go/updater"
	"github.com/thepwagner/action-update-go/updatertest"
)

func TestUpdater_Dependencies_Fixtures(t *testing.T) {
	cases := map[string][]updater.Dependency{
		"major": {
			{Path: "github.com/caarlos0/env/v5", Version: "v5.1.4"},
			{Path: "github.com/davecgh/go-spew", Version: "v1.1.1", Indirect: true},
		},
		"notinroot": {
			{Path: "github.com/pkg/errors", Version: "v0.8.0"},
		},
		"replace": {
			{Path: "github.com/thepwagner/errors", Version: "v0.8.0"},
		},
		"simple": {
			{Path: "github.com/pkg/errors", Version: "v0.8.0"},
			{Path: "github.com/sirupsen/logrus", Version: "v1.5.0"},
		},
		"vendor": {
			{Path: "github.com/pkg/errors", Version: "v0.8.0"},
		},
	}
	updatertest.DependenciesFixtures(t, func(t *testing.T, fixture string) updater.Updater {
		return updaterFromFixture(t, fixture)
	}, cases)
}
