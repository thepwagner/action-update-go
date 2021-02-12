package updater

import (
	"path"
	"strings"
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
	cleanPath := strings.ReplaceAll(update.Path, "https://", "")
	cleanPath = strings.ReplaceAll(cleanPath, "http://", "")
	cleanPath = strings.ReplaceAll(cleanPath, "#", "")
	cleanPath = strings.ReplaceAll(cleanPath, "{", "")
	cleanPath = strings.ReplaceAll(cleanPath, "}", "")
	return path.Join(branchPrefix, baseBranch, cleanPath, update.Next)
}

func (d DefaultUpdateBranchNamer) FormatBatch(baseBranch, batchName string) string {
	return path.Join(branchPrefix, baseBranch, batchName)
}
