package gomod_test

import (
	"fmt"
	"io/ioutil"
	"os"
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

func TestUpdater_UpdateAll_Simple(t *testing.T) {
	// Update and interrogate the logrus branch:
	r := updateAllInFixture(t, "simple")
	cfg, wt := checkoutBranchWithPrefix(t, r, "action-update-go/github.com/sirupsen/logrus/")

	// We expect 2 new branches: logrus and pkg/errors
	assert.Len(t, cfg.Branches, 3)

	// Logrus is upgraded, pkg/errors is not:
	goMod := worktreeFile(t, wt, gomod.GoModFn)
	assert.NotContains(t, goMod, "github.com/sirupsen/logrus v1.5.0", "logrus not updated")
	assert.Contains(t, goMod, "github.com/sirupsen/logrus")
	assert.Contains(t, goMod, "github.com/pkg/errors v0.8.0", "pkg/errors updated in wrong branch")
	goSum := worktreeFile(t, wt, "go.sum")
	assert.NotContains(t, goSum, "github.com/sirupsen/logrus v1.5.0", "go.sum not tidied")
	assert.Contains(t, goSum, "github.com/sirupsen/logrus")

	// No needless vendoring:
	_, err := wt.Filesystem.Stat(gomod.VendorModulesFn)
	assert.True(t, os.IsNotExist(err))
}

func TestUpdater_UpdateAll_Vendor(t *testing.T) {
	// Update and interrogate the logrus branch:
	r := updateAllInFixture(t, "vendor")
	cfg, wt := checkoutBranchWithPrefix(t, r, "action-update-go/github.com/sirupsen/logrus/")

	// We expect 1 new branches: logrus
	assert.Len(t, cfg.Branches, 2)

	// Logrus is upgraded:
	goMod := worktreeFile(t, wt, gomod.GoModFn)
	assert.NotContains(t, goMod, "github.com/sirupsen/logrus v1.5.0", "logrus not updated")
	assert.Contains(t, goMod, "github.com/sirupsen/logrus")

	modulesTxt := worktreeFile(t, wt, gomod.VendorModulesFn)
	assert.NotContains(t, modulesTxt, "github.com/sirupsen/logrus v1.5.0", "logrus not vendored")
}

func TestUpdater_UpdateAll_Major(t *testing.T) {
	// Update and interrogate the logrus branch:
	r := updateAllInFixture(t, "major")
	cfg, _ := checkoutBranchWithPrefix(t, r, "action-update-go/github.com/caarlos0/env/")

	// We expect 1 new branches: env
	assert.Len(t, cfg.Branches, 2)

}

func updateAllInFixture(t *testing.T, fixture string) *git.Repository {
	r := fixtureRepo(t, fixture)
	u, err := gomod.NewUpdater(r)
	require.NoError(t, err)
	err = u.UpdateAll("master")
	require.NoError(t, err)
	return r
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

func checkoutBranchWithPrefix(t *testing.T, r *git.Repository, prefix string) (*config.Config, *git.Worktree) {
	cfg, err := r.Config()
	require.NoError(t, err)
	branch := branchWithPrefix(cfg, prefix)
	require.NotEqualf(t, "", branch, "branch %q not found", prefix)

	wt, err := r.Worktree()
	require.NoError(t, err)
	err = wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branch),
		Force:  true,
	})
	require.NoError(t, err)
	return cfg, wt
}

func branchWithPrefix(cfg *config.Config, prefix string) string {
	for b := range cfg.Branches {
		if strings.HasPrefix(b, prefix) {
			return b
		}
	}
	return ""
}

func worktreeFile(t *testing.T, wt *git.Worktree, path string) string {
	goModFile, err := wt.Filesystem.Open(path)
	require.NoError(t, err)
	defer goModFile.Close()
	b, err := ioutil.ReadAll(goModFile)
	require.NoError(t, err)
	return string(b)
}
