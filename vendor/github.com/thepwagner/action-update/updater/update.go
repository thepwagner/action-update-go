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

// ExistingUpdate is a previously proposed update(s).
type ExistingUpdate struct {
	// Is this update still in a proposed state?
	Open bool
	// If not open, was this update accepted?
	Merged     bool
	BaseBranch string
	GroupName  string
	LastUpdate time.Time
	Updates    []Update
}
