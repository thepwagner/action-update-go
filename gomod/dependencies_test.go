package gomod_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thepwagner/action-update-go/updater"
)

func TestUpdater_Dependencies_Fixtures(t *testing.T) {
	// fixture name to expected dependencies
	cases := map[string][]updater.Dependency{
		"major": {
			{Path: "github.com/caarlos0/env/v5", Version: "v5.1.4"},
			{Path: "github.com/davecgh/go-spew", Version: "v1.1.1", Indirect: true},
		},
		"simple": {
			{Path: "github.com/pkg/errors", Version: "v0.8.0"},
			{Path: "github.com/sirupsen/logrus", Version: "v1.5.0"},
		},
		"vendor": {
			{Path: "github.com/pkg/errors", Version: "v0.8.0"},
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
