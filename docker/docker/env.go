package docker

import (
	"github.com/thepwagner/action-update/actions/updateaction"
	"github.com/thepwagner/action-update/updater"
)

type Environment struct {
	updateaction.Environment
}

func (c *Environment) NewUpdater(root string) updater.Updater {
	return NewUpdater(root)
}
