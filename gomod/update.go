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

type UpdatesByBranch map[string][]Update

func (p UpdatesByBranch) Add(baseBranch string, update Update) {
	existing := p[baseBranch]
	for _, u := range existing {
		if u.Path == update.Path && u.Next == update.Next {
			return
		}
	}
	p[baseBranch] = append(existing, update)
}
