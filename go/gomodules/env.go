package gomodules

import (
	"github.com/thepwagner/action-update/actions/updateaction"
	"github.com/thepwagner/action-update/updater"
)

type Environment struct {
	updateaction.Environment
	Tidy bool `env:"INPUT_TIDY" envDefault:"true"`
}

func (c *Environment) NewUpdater(root string) updater.Updater {
	return NewUpdater(root, WithTidy(c.Tidy))
}
