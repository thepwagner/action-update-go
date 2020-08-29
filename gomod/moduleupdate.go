package gomod

import (
	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/semver"
)

// ModuleUpdate is change in version
type ModuleUpdate struct {
	// Path of module being updated
	Path string
	// Previous module version
	Previous string
	// Next module version, if the upgrade is successful
	Next string
}

// Major returns true if the update changes major semver version
func (u ModuleUpdate) Major() bool {
	return semver.Major(u.Previous) != semver.Major(u.Next)
}
