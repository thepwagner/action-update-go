package repo_test

import (
	"errors"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thepwagner/action-update-go/repo"
)

const (
	branchName = "main"
	fileName   = "README.md"
)

var fileData = []byte{1, 2, 3, 4}

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

func TestNewGitRepo(t *testing.T) {
	gr := initGitRepo(t, plumbing.NewBranchReferenceName(branchName))
	assert.NotNil(t, gr)
}

func TestNewGitRepo_Dirty(t *testing.T) {
	// Repo with an uncommitted file:
	r := initRepo(t, plumbing.NewBranchReferenceName(branchName))
	wt, err := r.Worktree()
	require.NoError(t, err)
	_, err = ioutil.TempFile(wt.Filesystem.Root(), "dirty-git-")
	require.NoError(t, err)

	_, err = repo.NewGitRepo(r)
	assert.EqualError(t, err, "tree is dirty")
}

func TestGitRepo_Branch(t *testing.T) {
	gr := initGitRepo(t, plumbing.NewBranchReferenceName(branchName))
	assert.Equal(t, branchName, gr.Branch())
}

func TestGitRepo_BranchNotFound(t *testing.T) {
	gr := initGitRepo(t, "")
	assert.Equal(t, "", gr.Branch())
}

func TestGitRepo_SetBranch(t *testing.T) {
	gr := initGitRepo(t, plumbing.NewBranchReferenceName(branchName))
	err := gr.SetBranch(branchName)
	assert.NoError(t, err)
	assert.Equal(t, branchName, gr.Branch())
}

func TestGitRepo_SetBranchFromRemote(t *testing.T) {
	gr := initGitRepo(t, plumbing.NewRemoteReferenceName(repo.RemoteName, branchName))
	err := gr.SetBranch(branchName)
	assert.NoError(t, err)
	assert.Equal(t, branchName, gr.Branch())
}

func TestGitRepo_SetBranchNotFound(t *testing.T) {
	gr := initGitRepo(t, plumbing.NewBranchReferenceName(branchName))
	err := gr.SetBranch("not-" + branchName)
	assert.Equal(t, plumbing.ErrReferenceNotFound, errors.Unwrap(err))
}

func initGitRepo(t *testing.T, refName plumbing.ReferenceName) *repo.GitRepo {
	r := initRepo(t, refName)
	gr, err := repo.NewGitRepo(r)
	require.NoError(t, err)
	return gr
}

func initRepo(t *testing.T, refName plumbing.ReferenceName) *git.Repository {
	r, err := git.PlainInit(t.TempDir(), false)
	require.NoError(t, err)

	// If a ref is provided, initialize with const fileName+fileData
	if refName != "" {
		wt, err := r.Worktree()
		require.NoError(t, err)

		err = ioutil.WriteFile(filepath.Join(wt.Filesystem.Root(), fileName), fileData, 0644)
		require.NoError(t, err)
		_, err = wt.Add(fileName)
		require.NoError(t, err)
		commit, err := wt.Commit("initial", &git.CommitOptions{})
		require.NoError(t, err)

		err = r.Storer.SetReference(plumbing.NewHashReference(refName, commit))
		require.NoError(t, err)

		if refName.IsBranch() {
			err = wt.Checkout(&git.CheckoutOptions{
				Branch: refName,
			})
			require.NoError(t, err)
		}
	}

	return r
}
