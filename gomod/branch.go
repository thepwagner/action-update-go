package gomod

import "path"

const branchPrefix = "action-update-go"

type UpdateBranchNamer func(baseBranch string, update Update) string

var DefaultUpdateBranchNamer = func(baseBranch string, u Update) string {
	var branchPkg string
	if u.Major() {
		branchPkg = path.Dir(u.Path)
	} else {
		branchPkg = u.Path
	}
	return path.Join(branchPrefix, baseBranch, branchPkg, u.Next)
}
