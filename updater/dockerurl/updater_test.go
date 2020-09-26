package dockerurl_test

import (
	"fmt"

	"github.com/thepwagner/action-update-go/updater"
	"github.com/thepwagner/action-update-go/updater/dockerurl"
	"github.com/thepwagner/action-update-go/updatertest"
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
	dep     = updater.Dependency{Path: depPath, Version: previousVersion}
	update  = updater.Update{Path: depPath, Previous: previousVersion, Next: nextVersion}
)

func updaterFactory(opts ...dockerurl.UpdaterOpt) updatertest.Factory {
	return func(root string) updater.Updater { return dockerurl.NewUpdater(root, opts...) }
}
