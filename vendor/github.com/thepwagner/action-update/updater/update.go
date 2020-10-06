package updater

import (
	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/semver"
)

// Update is change included version to a specific module path
type Update struct {
	// Path of module being updated
	Path string `json:"path"`
	// Previous module version
	Previous string `json:"previous"`
	// Next module version
	Next string `json:"next"`
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
	existing.Open = addUpdate(existing.Open, update)
	p[baseBranch] = existing
}

func (p UpdatesByBranch) AddClosed(baseBranch string, update Update) {
	existing := p[baseBranch]
	existing.Closed = addUpdate(existing.Closed, update)
	p[baseBranch] = existing
}

func addUpdate(existing []Update, update Update) []Update {
	for _, u := range existing {
		if u.Path == update.Path && u.Next == update.Next {
			return existing
		}
	}
	return append(existing, update)
}

func (u Updates) OpenUpdate(update Update) string {
	return containsUpdate(u.Open, update)
}

func (u Updates) ClosedUpdate(update Update) string {
	return containsUpdate(u.Closed, update)
}

func (u Updates) Filter(update Update) string {
	if open := containsUpdate(u.Open, update); open != "" {
		return open
	}
	if closed := containsUpdate(u.Closed, update); closed != "" {
		return closed
	}
	return ""
}

func containsUpdate(updateList []Update, update Update) string {
	for _, u := range updateList {
		if u.Path != update.Path {
			continue
		}
		if semver.Compare(u.Next, update.Next) <= 0 {
			return u.Next
		}
	}
	return ""
}
