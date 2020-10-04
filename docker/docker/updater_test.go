package docker_test

import (
	"github.com/thepwagner/action-update-docker/docker"
	"github.com/thepwagner/action-update/updater"
)

type testFactory struct{}

func (u *testFactory) NewUpdater(root string) updater.Updater {
	return docker.NewUpdater(root)
}

func updaterFactory() updater.Factory {
	return &testFactory{}
}
