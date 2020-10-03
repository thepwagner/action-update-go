package updater_test

import (
	"github.com/thepwagner/action-update-go/updater/docker"
	"github.com/thepwagner/action-update-go/updatertest"
	updater2 "github.com/thepwagner/action-update/updater"
)

func updaterFactory() updatertest.Factory {
	return func(root string) updater2.Updater { return docker.NewUpdater(root) }
}
