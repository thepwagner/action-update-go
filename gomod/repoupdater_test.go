package gomod_test

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/otiai10/copy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thepwagner/action-update-go/gomod"
	gitrepo "github.com/thepwagner/action-update-go/repo"
)

const baseBranchCount = 2

func TestUpdater_UpdateAll_Simple(t *testing.T) {
	// Update and interrogate the logrus branch:
	r := updateAllInFixture(t, "simple")
	branches, wt := checkoutBranchWithPrefix(t, r, "action-update-go/master/github.com/sirupsen/logrus/")

	// We expect 2 new branches: logrus and pkg/errors
	assert.Len(t, branches, baseBranchCount+2)

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

func TestUpdater_UpdateAll_MultiBranch(t *testing.T) {
	upstream, downstream := fixtureRepos(t, "simple")
	r, err := gitrepo.NewGitRepo(downstream)
	require.NoError(t, err)
	u, err := gomod.NewRepoUpdater(r)
	require.NoError(t, err)

	ctx := context.Background()
	for _, b := range []string{"master", "main"} {
		err = u.UpdateAll(ctx, b)
		require.NoError(t, err)
	}

	branches := iterateBranches(t, upstream)
	// We expect 4 new branches: logrus and pkg/errors for each base branch
	assert.Len(t, branches, baseBranchCount+4)
}

func updateAllInFixture(t *testing.T, fixture string) *git.Repository {
	upstream, downstream := fixtureRepos(t, fixture)
	r, err := gitrepo.NewGitRepo(downstream)
	require.NoError(t, err)
	u, err := gomod.NewRepoUpdater(r)
	require.NoError(t, err)
	err = u.UpdateAll(context.Background(), "master")
	require.NoError(t, err)
	return upstream
}

func fixtureRepos(t *testing.T, fixture string) (upstream, downstream *git.Repository) {
	// Init upstream in a tempdir:
	upstreamRepo := t.TempDir()
	t.Logf("upstream dir: %s", upstreamRepo)
	upstream, err := git.PlainInit(upstreamRepo, false)
	require.NoError(t, err)

	// Fill with files from the fixture:
	err = copy.Copy(fmt.Sprintf("../fixtures/%s", fixture), upstreamRepo)
	require.NoError(t, err)

	// Add as initial commit:
	wt, err := upstream.Worktree()
	require.NoError(t, err)
	err = wt.AddGlob(".")
	require.NoError(t, err)
	commit, err := wt.Commit("initial", &git.CommitOptions{
		Author: &object.Signature{
			Name:  "test",
			Email: "test@test.com",
		},
	})
	require.NoError(t, err)

	otherBranch := plumbing.NewHashReference(plumbing.NewBranchReferenceName("main"), commit)
	err = upstream.Storer.SetReference(otherBranch)
	require.NoError(t, err)

	err = upstream.CreateBranch(&config.Branch{
		Name: "main",
	})
	require.NoError(t, err)

	downstreamRepo := t.TempDir()
	t.Logf("downstream dir: %s", downstreamRepo)
	downstream, err = git.PlainClone(downstreamRepo, false, &git.CloneOptions{
		URL: upstreamRepo,
	})
	require.NoError(t, err)
	return upstream, downstream
}

func checkoutBranchWithPrefix(t *testing.T, r *git.Repository, prefix string) (map[string]struct{}, *git.Worktree) {
	branches := iterateBranches(t, r)
	var branch string
	for b := range branches {
		if strings.HasPrefix(b, prefix) {
			branch = b
			break
		}
	}
	require.NotEqualf(t, "", branch, "branch %q not found", prefix)

	wt, err := r.Worktree()
	require.NoError(t, err)
	err = wt.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(branch),
		Force:  true,
	})
	require.NoError(t, err)
	return branches, wt
}

func iterateBranches(t *testing.T, r *git.Repository) map[string]struct{} {
	ret := map[string]struct{}{}
	branchIter, err := r.Branches()
	require.NoError(t, err)
	for {
		next, err := branchIter.Next()
		if err == io.EOF {
			break
		}
		require.NoError(t, err)
		ret[next.Name().Short()] = struct{}{}
	}
	return ret
}

func worktreeFile(t *testing.T, wt *git.Worktree, path string) string {
	goModFile, err := wt.Filesystem.Open(path)
	require.NoError(t, err)
	defer goModFile.Close()
	b, err := ioutil.ReadAll(goModFile)
	require.NoError(t, err)
	return string(b)
}
