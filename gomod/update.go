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

type UpdatesByBranch map[string]Updates

type Updates struct {
	// Open updates are waiting user action
	Open []Update
	// Closed updates are acted negatively on by the user (e.g. PR closed without merge)
	Closed []Update
}

func (p UpdatesByBranch) AddOpen(baseBranch string, update Update) {
	existing := p[baseBranch]
	for _, u := range existing.Open {
		if u.Path == update.Path && u.Next == update.Next {
			return
		}
	}
	existing.Open = append(existing.Open, update)
	p[baseBranch] = existing
}
