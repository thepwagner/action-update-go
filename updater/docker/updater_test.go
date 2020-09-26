package docker_test

import (
	"github.com/thepwagner/action-update-go/updater"
	"github.com/thepwagner/action-update-go/updater/docker"
	"github.com/thepwagner/action-update-go/updatertest"
)

func updaterFactory() updatertest.Factory {
	return func(root string) updater.Updater { return docker.NewUpdater(root) }
}
