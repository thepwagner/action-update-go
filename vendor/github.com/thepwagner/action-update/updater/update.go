package updater

import "time"

// Update specifies changes to the version of a specific module path
type Update struct {
	// Path of module being updated
	Path string `json:"path"`
	// Previous module version
	Previous string `json:"previous"`
	// Next module version
	Next string `json:"next"`
}

type UpdateGroup struct {
	Name    string   `json:"name,omitempty"`
	Updates []Update `json:"updates"`
}

func NewUpdateGroup(name string, updates ...Update) UpdateGroup {
	return UpdateGroup{Name: name, Updates: updates}
}

// ExistingUpdate is a previously proposed update(s).
type ExistingUpdate struct {
	// Is this update still in a proposed state?
	Open bool
	// If not open, was this update accepted?
	Merged     bool
	BaseBranch string
	LastUpdate time.Time
	Group      UpdateGroup
}

type ExistingUpdates []ExistingUpdate

func (e ExistingUpdates) LatestGroupUpdate(group string) (latest time.Time) {
	for _, u := range e {
		if u.Group.Name != group {
			continue
		}
		if u.LastUpdate.After(latest) {
			latest = u.LastUpdate
		}
	}
	return
}
