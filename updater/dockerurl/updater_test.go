package dockerurl_test

import (
	"fmt"

	"github.com/thepwagner/action-update-go/updater"
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
