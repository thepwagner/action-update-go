package repo_test

import (
	"context"
	"errors"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thepwagner/action-update/repo"
	"github.com/thepwagner/action-update/updater"
)

const (
	branchName   = "main"
	fileName     = "README.md"
	updateBranch = "action-update-go/main/github.com/foo/bar/v1.0.0"
)

var fileData = []byte{1, 2, 3, 4}

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

func TestGitRepo_Branch_NotFound(t *testing.T) {
	gr := initGitRepo(t, "")
	assert.Equal(t, "", gr.Branch())
}

func TestGitRepo_SetBranch(t *testing.T) {
	gr := initGitRepo(t, plumbing.NewBranchReferenceName(branchName))
	err := gr.SetBranch(branchName)
	assert.NoError(t, err)
	assert.Equal(t, branchName, gr.Branch())
}

func TestGitRepo_SetBranch_FromRemote(t *testing.T) {
	gr := initGitRepo(t, plumbing.NewRemoteReferenceName(repo.RemoteName, branchName))
	err := gr.SetBranch(branchName)
	assert.NoError(t, err)
	assert.Equal(t, branchName, gr.Branch())
}

func TestGitRepo_SetBranch_NotFound(t *testing.T) {
	gr := initGitRepo(t, plumbing.NewBranchReferenceName(branchName))
	err := gr.SetBranch("not-" + branchName)
	assert.Equal(t, plumbing.ErrReferenceNotFound, errors.Unwrap(err))
}

func TestGitRepo_NewBranch(t *testing.T) {
	gr := initGitRepo(t, plumbing.NewBranchReferenceName(branchName))
	err := gr.NewBranch(branchName, updateBranch)
	assert.NoError(t, err)
	assert.Equal(t, updateBranch, gr.Branch())
}

func TestGitRepo_NewBranch_FromRemote(t *testing.T) {
	gr := initGitRepo(t, plumbing.NewRemoteReferenceName(repo.RemoteName, branchName))
	err := gr.NewBranch(branchName, updateBranch)
	assert.NoError(t, err)
	assert.Equal(t, updateBranch, gr.Branch())
}

func TestGitRepo_Push(t *testing.T) {
	gr := initGitRepo(t, plumbing.NewRemoteReferenceName(repo.RemoteName, branchName))
	err := gr.NewBranch(branchName, updateBranch)
	require.NoError(t, err)
	tmpFile := addTempFile(t, gr)

	err = gr.Push(context.Background(), updater.Update{
		Path: "github.com/test",
		Next: "v1.0.0",
	})
	require.NoError(t, err)

	// Re-open repo and get log from the update branch:
	r, err := git.PlainOpen(gr.Root())
	require.NoError(t, err)
	branchRef, err := r.Reference(plumbing.NewBranchReferenceName(updateBranch), true)
	require.NoError(t, err)
	log, err := r.Log(&git.LogOptions{
		From: branchRef.Hash(),
	})
	require.NoError(t, err)
	defer log.Close()

	// Verify commit:
	commit, err := log.Next()
	require.NoError(t, err)
	t.Logf("inspecting commit %s", commit.Hash)
	assert.Equal(t, "github.com/test@v1.0.0", commit.Message)
	assert.Equal(t, repo.DefaultGitIdentity.Name, commit.Author.Name)
	assert.Equal(t, repo.DefaultGitIdentity.Email, commit.Author.Email)

	// File was added to tree with expected contents:
	tree, err := commit.Tree()
	require.NoError(t, err)
	f, err := tree.File(tmpFile)
	require.NoError(t, err)
	fContents, err := f.Contents()
	require.NoError(t, err)
	assert.Equal(t, tmpFile, fContents)
}

func TestGitRepo_Push_WithRemote(t *testing.T) {
	// Initialize a repo, clone it, then use the clone:
	upstream := initRepo(t, plumbing.NewBranchReferenceName(branchName))
	upstreamWt, err := upstream.Worktree()
	require.NoError(t, err)
	downstream, err := git.PlainClone(t.TempDir(), false, &git.CloneOptions{
		URL: upstreamWt.Filesystem.Root(),
	})
	require.NoError(t, err)

	gr, err := repo.NewGitRepo(downstream)
	require.NoError(t, err)
	err = gr.NewBranch(branchName, updateBranch)
	require.NoError(t, err)
	addTempFile(t, gr)

	err = gr.Push(context.Background(), updater.Update{
		Path: "github.com/test",
		Next: "v1.0.0",
	})
	require.NoError(t, err)

	// Branch was pushed to upstream repo:
	_, err = upstream.Reference(plumbing.NewBranchReferenceName(updateBranch), true)
	assert.NoError(t, err)
}

func TestGitRepo_Fetch(t *testing.T) {
	// Initialize a repo, clone it:
	upstream := initRepo(t, plumbing.NewBranchReferenceName(branchName))
	upstreamWt, err := upstream.Worktree()
	require.NoError(t, err)
	downstream, err := git.PlainClone(t.TempDir(), false, &git.CloneOptions{
		URL: upstreamWt.Filesystem.Root(),
	})
	require.NoError(t, err)

	// Create a branch in the upstream repo:
	newBranch := "test-1234"
	upstreamRepo, err := repo.NewGitRepo(upstream)
	require.NoError(t, err)
	err = upstreamRepo.NewBranch(branchName, newBranch)
	require.NoError(t, err)

	// Fetch from the downstream repo, the branch is usable:
	downstreamRepo, err := repo.NewGitRepo(downstream)
	require.NoError(t, err)
	err = downstreamRepo.Fetch(context.Background(), newBranch)
	require.NoError(t, err)
	err = downstreamRepo.NewBranch(newBranch, updateBranch)
	require.NoError(t, err)
}

func addTempFile(t *testing.T, gr *repo.GitRepo) string {
	f, err := ioutil.TempFile(gr.Root(), "my-awesome-file-")
	require.NoError(t, err)
	fn := filepath.Base(f.Name())
	_, err = f.WriteString(fn)
	require.NoError(t, err)
	require.NoError(t, f.Close())
	return fn
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
		commit, err := wt.Commit("initial", &git.CommitOptions{
			Author: &object.Signature{
				Name:  "test",
				Email: "test@test.com",
			},
		})
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
