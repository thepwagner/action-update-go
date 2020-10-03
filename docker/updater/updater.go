package updater

import (
	"context"

	updater2 "github.com/thepwagner/action-update/updater"
)

type Updater struct {
	root string
}

var _ updater2.Updater = (*Updater)(nil)

func NewUpdater(root string) *Updater {
	return &Updater{root: root}
}

func (u *Updater) Check(ctx context.Context, dependency updater2.Dependency) (*updater2.Update, error) {
	panic("implement me")
}

func (u *Updater) ApplyUpdate(ctx context.Context, update updater2.Update) error {
	panic("implement me")
}
