package docker

import (
	"context"

	"github.com/thepwagner/action-update/updater"
)

type Updater struct {
	root string
}

var _ updater.Updater = (*Updater)(nil)

func NewUpdater(root string) *Updater {
	return &Updater{root: root}
}

var _ updater.Updater = (*Updater)(nil)

func (u *Updater) Check(ctx context.Context, dependency updater.Dependency) (*updater.Update, error) {
	panic("implement me")
}

func (u *Updater) ApplyUpdate(ctx context.Context, update updater.Update) error {
	panic("implement me")
}
