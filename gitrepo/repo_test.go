package gitrepo_test

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thepwagner/action-update-go/gitrepo"
)

const (
	branchName = "main"
	fileName   = "README.md"
)

var fileData = []byte{1, 2, 3, 4}

func TestNewSharedRepo(t *testing.T) {
	sharedRepo := initSharedRepo(t, plumbing.NewBranchReferenceName(branchName))
	assert.NotNil(t, sharedRepo)
}

func TestNewSharedRepo_Dirty(t *testing.T) {
	// Repo with an uncommitted file:
	repo := initRepo(t, plumbing.NewBranchReferenceName(branchName))
	wt, err := repo.Worktree()
	require.NoError(t, err)
	_, err = ioutil.TempFile(wt.Filesystem.Root(), "dirty-git-")
	require.NoError(t, err)

	_, err = gitrepo.NewSharedRepo(repo)
	assert.EqualError(t, err, "tree is dirty")
}

func TestSharedRepo_ReadFile(t *testing.T) {
	sharedRepo := initSharedRepo(t, plumbing.NewBranchReferenceName(branchName))
	b, err := sharedRepo.ReadFile(branchName, fileName)
	require.NoError(t, err)
	assert.Equal(t, fileData, b)
}

func TestSharedRepo_ReadFile_BranchFoundRemote(t *testing.T) {
	sharedRepo := initSharedRepo(t, plumbing.NewRemoteReferenceName(gitrepo.RemoteName, branchName))
	b, err := sharedRepo.ReadFile(branchName, fileName)
	require.NoError(t, err)
	assert.Equal(t, fileData, b)
}

func TestSharedRepo_ReadFile_BranchNotFound(t *testing.T) {
	sharedRepo := initSharedRepo(t, plumbing.NewBranchReferenceName(branchName))
	_, err := sharedRepo.ReadFile("not-"+branchName, fileName)
	assert.Equal(t, plumbing.ErrReferenceNotFound, errors.Unwrap(err))
}

func TestSharedRepo_ReadFile_FileNotFound(t *testing.T) {
	repo := initRepo(t, plumbing.NewBranchReferenceName(branchName))
	sharedRepo, err := gitrepo.NewSharedRepo(repo)
	require.NoError(t, err)
	_, err = sharedRepo.ReadFile(branchName, "not-"+fileName)
	assert.True(t, os.IsNotExist(err))
}

func TestSharedRepo_NewSandbox(t *testing.T) {
	repo := initRepo(t, plumbing.NewBranchReferenceName(branchName))
	sharedRepo, err := gitrepo.NewSharedRepo(repo)
	require.NoError(t, err)

	sbx, err := sharedRepo.NewSandbox(branchName, "my-awesome-branch")
	require.NoError(t, err)
	assert.NotNil(t, sbx)
}

func initSharedRepo(t *testing.T, refName plumbing.ReferenceName) *gitrepo.SharedRepo {
	repo := initRepo(t, refName)
	sharedRepo, err := gitrepo.NewSharedRepo(repo)
	require.NoError(t, err)
	return sharedRepo
}

func initRepo(t *testing.T, refName plumbing.ReferenceName) *git.Repository {
	repo, err := git.PlainInit(t.TempDir(), false)
	require.NoError(t, err)
	wt, err := repo.Worktree()
	require.NoError(t, err)

	err = ioutil.WriteFile(filepath.Join(wt.Filesystem.Root(), fileName), fileData, 0644)
	require.NoError(t, err)
	_, err = wt.Add(fileName)
	require.NoError(t, err)
	commit, err := wt.Commit("initial", &git.CommitOptions{})
	require.NoError(t, err)

	err = repo.Storer.SetReference(plumbing.NewHashReference(refName, commit))
	require.NoError(t, err)
	return repo
}
