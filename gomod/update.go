package gomod

import (
	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/semver"
)

// Update is change in version to a specific module path
type Update struct {
	// Path of module being updated
	Path string
	// Previous module version
	Previous string
	// Next module version
	Next string
}

// Major returns true if the update changes major version
func (u Update) Major() bool {
	return semver.Major(u.Previous) != semver.Major(u.Next)
}