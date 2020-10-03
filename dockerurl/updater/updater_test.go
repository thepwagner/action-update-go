package updater_test

import (
	"fmt"

	"github.com/thepwagner/action-update-go/updater/dockerurl"
	"github.com/thepwagner/action-update-go/updatertest"
	updater2 "github.com/thepwagner/action-update/updater"
)

//go:generate mockery --outpkg dockerurl_test --output . --testonly --name repoClient --structname mockRepoClient --filename mockrepoclient_test.go

const (
	previousVersion = "v1.4.0"
	nextVersion     = "v1.4.1"

	depOwner    = "containerd"
	depRepoName = "containerd"
)

var (
	depPath = fmt.Sprintf("github.com/%s/%s/releases", depOwner, depRepoName)
	dep     = updater2.Dependency{Path: depPath, Version: previousVersion}
	update  = updater2.Update{Path: depPath, Previous: previousVersion, Next: nextVersion}
)

func updaterFactory(opts ...dockerurl.UpdaterOpt) updatertest.Factory {
	return func(root string) updater2.Updater { return dockerurl.NewUpdater(root, opts...) }
}
