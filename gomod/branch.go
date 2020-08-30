package gomod

import (
	"path"
	"strings"
)

const branchPrefix = "action-update-go"

type UpdateBranchNamer interface {
	Format(baseBranch string, update Update) string
	Parse(string) (baseBranch string, update *Update)
}

type DefaultUpdateBranchNamer struct{}

func (d DefaultUpdateBranchNamer) Format(baseBranch string, update Update) string {
	return path.Join(branchPrefix, baseBranch, update.Path, update.Next)
}

func (d DefaultUpdateBranchNamer) Parse(branch string) (baseBranch string, update *Update) {
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
