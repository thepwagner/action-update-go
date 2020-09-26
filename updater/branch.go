package updater

import (
	"path"
	"strings"
)

const branchPrefix = "action-update-go"

// UpdateBranchNamer names branches for proposed updates.
type UpdateBranchNamer interface {
	Format(baseBranch string, update Update) string
	FormatBatch(baseBranch, batchName string) string
	Parse(string) (baseBranch string, update *Update)
}

type DefaultUpdateBranchNamer struct{}

var _ UpdateBranchNamer = (*DefaultUpdateBranchNamer)(nil)

func (d DefaultUpdateBranchNamer) Format(baseBranch string, update Update) string {
	return path.Join(branchPrefix, baseBranch, update.Path, update.Next)
}

func (d DefaultUpdateBranchNamer) FormatBatch(baseBranch, batchName string) string {
	return path.Join(branchPrefix, baseBranch, batchName)
}

func (d DefaultUpdateBranchNamer) Parse(branch string) (baseBranch string, u *Update) {
	branchSplit := strings.Split(branch, "/")
	if len(branchSplit) < 4 || branchSplit[0] != branchPrefix {
		return "", nil
	}
	versPos := len(branchSplit) - 1
	return branchSplit[1], &Update{
		Path: path.Join(branchSplit[2:versPos]...),
		Next: branchSplit[versPos],
	}
}
