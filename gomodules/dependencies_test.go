package gomodules_test

import (
	"testing"

	"github.com/thepwagner/action-update/updater"
	"github.com/thepwagner/action-update/updatertest"
)

func TestUpdater_Dependencies_Fixtures(t *testing.T) {
	cases := map[string][]updater.Dependency{
		"gopkg": {
			{Path: "gopkg.in/yaml.v1", Version: "v1.0.0-20140924161607-9f9df34309c0"},
		},
		"major": {
			{Path: "github.com/caarlos0/env/v5", Version: "v5.1.4"},
			{Path: "github.com/davecgh/go-spew", Version: "v1.1.1", Indirect: true},
		},
		"multimodule": {
			{Path: "github.com/pkg/errors", Version: "v0.8.0"},
			{Path: "github.com/sirupsen/logrus", Version: "v1.5.0"},
		},
		"multimodule/common": {
			{Path: "github.com/sirupsen/logrus", Version: "v1.5.0"},
		},
		"multimodule/cmd": {
			{Path: "github.com/pkg/errors", Version: "v0.8.0"},
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

	updatertest.DependenciesFixtures(t, updaterFactory(), cases)
}
