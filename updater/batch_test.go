package updater_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thepwagner/action-update-go/updater"
)

func TestGroupDependencies_NoBatch(t *testing.T) {
	dep := updater.Dependency{Path: mockPath}
	b := updater.GroupDependencies(nil, []updater.Dependency{dep})

	assert.Equal(t, map[string][]updater.Dependency{
		"": {dep},
	}, b)
}

func TestGroupDependencies_Batch(t *testing.T) {
	dep := updater.Dependency{Path: mockPath}
	b := updater.GroupDependencies(
		map[string][]string{
			"not-awesome":  {"prefix-does-not-match"},
			"semi-awesome": {"github.com/"}, // match, but there is a longer prefix
			"awesome":      {"github.com/foo"},
		}, []updater.Dependency{dep})

	assert.Equal(t, map[string][]updater.Dependency{
		"awesome": {dep},
	}, b)
}
