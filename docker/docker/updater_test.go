package docker_test

import (
	"github.com/thepwagner/action-update-docker/docker"
	"github.com/thepwagner/action-update/updater"
)

func updaterFactory() updater.Factory {
	return func(root string) updater.Updater { return docker.NewUpdater(root) }
}
