package gomod

import (
	"fmt"
	"path"

	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/semver"
	"github.com/go-git/go-git/v5/plumbing"
)

// ModuleUpdate is change in version
type ModuleUpdate struct {
	// Base is the
	Base *plumbing.Reference
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

// BranchName returns the branch name for this update.
func (u ModuleUpdate) BranchName() string {
	var branchPkg string
	if u.Major() {
		branchPkg = path.Dir(u.Path)
	} else {
		branchPkg = u.Path
	}
	return fmt.Sprintf("action-update-go/%s/%s/%s", u.Base.Name().Short(), branchPkg, u.Next)
}
