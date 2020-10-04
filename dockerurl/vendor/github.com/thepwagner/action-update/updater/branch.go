package updater

import (
	"path"
)

const branchPrefix = "action-update-go"

// UpdateBranchNamer names branches for proposed updates.
type UpdateBranchNamer interface {
	// Format generates branch name for an update
	Format(baseBranch string, update Update) string
	// Format generates branch name for batch of updates
	FormatBatch(baseBranch, batchName string) string
}

type DefaultUpdateBranchNamer struct{}

var _ UpdateBranchNamer = (*DefaultUpdateBranchNamer)(nil)

func (d DefaultUpdateBranchNamer) Format(baseBranch string, update Update) string {
	return path.Join(branchPrefix, baseBranch, update.Path, update.Next)
}

func (d DefaultUpdateBranchNamer) FormatBatch(baseBranch, batchName string) string {
	return path.Join(branchPrefix, baseBranch, batchName)
}
