package gomod_test

import (
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	copy2 "github.com/otiai10/copy"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thepwagner/action-update-go/gomod"
)

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

func TestUpdater_UpdateAll(t *testing.T) {
	r := fixtureRepo(t, "simple")
	u, err := gomod.NewUpdater(r)
	require.NoError(t, err)

	err = u.UpdateAll("master")
	require.NoError(t, err)

	// We expect a branch has been created containing updates to logrus:
	cfg, err := r.Config()
	require.NoError(t, err)
	assert.Len(t, cfg.Branches, 2)
	var logrusBranch string
	for b := range cfg.Branches {
		if strings.HasPrefix(b, "action-update-go/github.com/sirupsen/logrus/") {
			logrusBranch = b
			break
		}
	}
	require.NotEqual(t, "", logrusBranch, "logrus update branch not found")

	wt, err := r.Worktree()
	require.NoError(t, err)
	err = wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(logrusBranch),
		Force:  true,
	})
	require.NoError(t, err)

	goMod := worktreeFile(t, wt, "go.mod")
	assert.NotContains(t, goMod, "github.com/sirupsen/logrus v1.5.0", "logrus not updated")
	assert.Contains(t, goMod, "github.com/sirupsen/logrus")

	goSum := worktreeFile(t, wt, "go.sum")
	assert.NotContains(t, goSum, "github.com/sirupsen/logrus v1.5.0", "go.sum not tidied")
	assert.Contains(t, goSum, "github.com/sirupsen/logrus")
}

func worktreeFile(t *testing.T, wt *git.Worktree, path string) string {
	goModFile, err := wt.Filesystem.Open(path)
	require.NoError(t, err)
	defer goModFile.Close()
	b, err := ioutil.ReadAll(goModFile)
	require.NoError(t, err)
	return string(b)
}

func fixtureRepo(t *testing.T, fixture string) *git.Repository {
	// Init repo in a tempdir:
	tmpDir := t.TempDir()
	t.Logf("repo dir: %s", tmpDir)
	repo, err := git.PlainInit(tmpDir, false)
	require.NoError(t, err)

	// Fill with files from the fixture:
	err = copy2.Copy(fmt.Sprintf("../fixtures/%s", fixture), tmpDir)
	require.NoError(t, err)

	// Add as initial commit:
	wt, err := repo.Worktree()
	require.NoError(t, err)
	err = wt.AddGlob(".")
	require.NoError(t, err)
	_, err = wt.Commit("initial", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "test",
			Email: "test@test.com",
		},
	})
	require.NoError(t, err)
	err = repo.CreateBranch(&config.Branch{
		Name: "main",
	})
	require.NoError(t, err)
	return repo
}
