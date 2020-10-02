package updater_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thepwagner/action-update-go/updater"
)

func TestDefaultUpdateBranchNamer_Format(t *testing.T) {
	const baseBranch = "main"
	branchNamer := updater.DefaultUpdateBranchNamer{}

	cases := []struct {
		branch string
		update updater.Update
	}{
		{
			update: updater.Update{
				Path: "github.com/foo/bar",
				Next: "v1.2.3",
			},
			branch: "action-update-go/main/github.com/foo/bar/v1.2.3",
		},
		{
			update: updater.Update{
				Path:     "github.com/foo/bar/v2",
				Previous: "v2.0.0",
				Next:     "v3.0.0",
			},
			branch: "action-update-go/main/github.com/foo/bar/v2/v3.0.0",
		},
		{
			update: updater.Update{
				Path:     "github.com/foo/bar",
				Previous: "v2.0.0",
				Next:     "v3.0.0",
			},
			branch: "action-update-go/main/github.com/foo/bar/v3.0.0",
		},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%v", c.branch), func(t *testing.T) {
			branch := branchNamer.Format(baseBranch, c.update)
			assert.Equal(t, c.branch, branch)
		})
	}
}
