package gomod_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thepwagner/action-update-go/gomod"
)

func TestDefaultUpdateBranchName(t *testing.T) {
	const baseBranch = "main"
	branchNamer := gomod.DefaultUpdateBranchNamer{}

	cases := []struct {
		branch string
		update *gomod.Update
	}{
		{
			branch: "my-awesome-branch",
			update: nil,
		},
		{
			update: &gomod.Update{
				Path: "github.com/foo/bar",
				Next: "v1.2.3",
			},
			branch: "action-update-go/main/github.com/foo/bar/v1.2.3",
		},
		{
			update: &gomod.Update{
				Path:     "github.com/foo/bar/v2",
				Previous: "v2.0.0",
				Next:     "v3.0.0",
			},
			branch: "action-update-go/main/github.com/foo/bar/v2/v3.0.0",
		},
		{
			update: &gomod.Update{
				Path:     "github.com/foo/bar",
				Previous: "v2.0.0",
				Next:     "v3.0.0",
			},
			branch: "action-update-go/main/github.com/foo/bar/v3.0.0",
		},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%v", c.branch), func(t *testing.T) {
			base, update := branchNamer.Parse(c.branch)

			if c.update != nil {
				assert.Equal(t, baseBranch, base)
				assert.Equal(t, c.update.Path, update.Path)
				assert.Equal(t, c.update.Next, update.Next)

				formatted := branchNamer.Format(baseBranch, *c.update)
				assert.Equal(t, c.branch, formatted)
			} else {
				assert.Equal(t, "", base)
				assert.Equal(t, "", base)
				assert.Nil(t, c.update)
			}
		})
	}
}
