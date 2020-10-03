package updater_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	updater2 "github.com/thepwagner/action-update/updater"
)

func TestGroupDependencies_NoBatch(t *testing.T) {
	dep := updater2.Dependency{Path: mockPath}
	b := updater2.GroupDependencies(nil, []updater2.Dependency{dep})

	assert.Equal(t, map[string][]updater2.Dependency{
		"": {dep},
	}, b)
}

func TestGroupDependencies_Batch(t *testing.T) {
	dep := updater2.Dependency{Path: mockPath}
	b := updater2.GroupDependencies(
		map[string][]string{
			"not-awesome":  {"prefix-does-not-match"},
			"semi-awesome": {"github.com/"}, // match, but there is a longer prefix
			"awesome":      {"github.com/foo"},
		}, []updater2.Dependency{dep})

	assert.Equal(t, map[string][]updater2.Dependency{
		"awesome": {dep},
	}, b)
}
