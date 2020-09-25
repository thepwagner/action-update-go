package repo

import (
	"path"
	"strings"

	"github.com/thepwagner/action-update-go/updater"
)

const branchPrefix = "action-update-go"

// UpdateBranchNamer names branches for proposed updates.
type UpdateBranchNamer interface {
	Format(baseBranch string, update updater.Update) string
	Parse(string) (baseBranch string, update *updater.Update)
}

type DefaultUpdateBranchNamer struct{}

var _ UpdateBranchNamer = (*DefaultUpdateBranchNamer)(nil)

func (d DefaultUpdateBranchNamer) Format(baseBranch string, update updater.Update) string {
	return path.Join(branchPrefix, baseBranch, update.Path, update.Next)
}

func (d DefaultUpdateBranchNamer) Parse(branch string) (baseBranch string, u *updater.Update) {
	branchSplit := strings.Split(branch, "/")
	if len(branchSplit) < 4 || branchSplit[0] != branchPrefix {
		return "", nil
	}
	versPos := len(branchSplit) - 1
	return branchSplit[1], &updater.Update{
		Path: path.Join(branchSplit[2:versPos]...),
		Next: branchSplit[versPos],
	}
}
