package gomodules

import (
	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/semver"
	"github.com/thepwagner/action-update/updater"
)

type Updater struct {
	root string

	// MajorVersion attempts major package versions, e.g. github.com/foo/bar/v2 -> github.com/foo/bar/v3
	MajorVersions bool
	// Tidy toggles `go mod tidy` after an update
	Tidy bool
}

var _ updater.Updater = (*Updater)(nil)

func NewUpdater(root string, opts ...UpdaterOpt) *Updater {
	u := &Updater{
		root: root,

		MajorVersions: true,
		Tidy:          true,
	}
	for _, opt := range opts {
		opt(u)
	}
	return u
}

type UpdaterOpt func(*Updater)

func WithTidy(tidy bool) UpdaterOpt {
	return func(u *Updater) {
		u.Tidy = tidy
	}
}

func WithMajorVersions(major bool) UpdaterOpt {
	return func(u *Updater) {
		u.MajorVersions = major
	}
}

const (
	GoModFn         = "go.mod"
	GoSumFn         = "go.sum"
	VendorModulesFn = "vendor/modules.txt"
)

func MajorPkg(u updater.Update) bool {
	return semver.Major(u.Previous) != semver.Major(u.Next) && pathMajorVersionRE.MatchString(u.Path)
}
