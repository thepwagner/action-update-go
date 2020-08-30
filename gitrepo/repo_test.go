package gitrepo_test

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
	"github.com/thepwagner/action-update-go/gitrepo"
)

const (
	branchName = "main"
	fileName   = "README.md"
)

var fileData = []byte{1, 2, 3, 4}

func init() {
	logrus.SetLevel(logrus.DebugLevel)
}

func TestSingleTreeRepo(t *testing.T) {
	singleTree := initSingleTree(t, plumbing.NewBranchReferenceName(branchName))
	assert.NotNil(t, singleTree)
}

func TestSingleTreeRepo_Dirty(t *testing.T) {
	// Repo with an uncommitted file:
	repo := initRepo(t, plumbing.NewBranchReferenceName(branchName))
	wt, err := repo.Worktree()
	require.NoError(t, err)
	_, err = ioutil.TempFile(wt.Filesystem.Root(), "dirty-git-")
	require.NoError(t, err)

	_, err = gitrepo.NewSingleTreeRepo(repo)
	assert.EqualError(t, err, "tree is dirty")
}

func TestSingleTreeRepo_Branch(t *testing.T) {
	singleTree := initSingleTree(t, plumbing.NewBranchReferenceName(branchName))
	assert.Equal(t, branchName, singleTree.Branch())
}

func TestSingleTreeRepo_BranchNotFound(t *testing.T) {
	singleTree := initSingleTree(t, "")
	assert.Equal(t, "", singleTree.Branch())
}

func TestSingleTreeRepo_SetBranch(t *testing.T) {
	singleTree := initSingleTree(t, plumbing.NewBranchReferenceName(branchName))
	err := singleTree.SetBranch(branchName)
	assert.NoError(t, err)
	assert.Equal(t, branchName, singleTree.Branch())
}

func TestSingleTreeRepo_SetBranchFromRemote(t *testing.T) {
	singleTree := initSingleTree(t, plumbing.NewRemoteReferenceName(gitrepo.RemoteName, branchName))
	err := singleTree.SetBranch(branchName)
	assert.NoError(t, err)
	assert.Equal(t, branchName, singleTree.Branch())
}

func TestSingleTreeRepo_SetBranchNotFound(t *testing.T) {
	singleTree := initSingleTree(t, plumbing.NewBranchReferenceName(branchName))
	err := singleTree.SetBranch("not-" + branchName)
	assert.Equal(t, plumbing.ErrReferenceNotFound, errors.Unwrap(err))
}

//func TestSharedRepo_ReadFile(t *testing.T) {
//	sharedRepo := initSingleTree(t, plumbing.NewBranchReferenceName(branchName))
//	b, err := sharedRepo.ReadFile(branchName, fileName)
//	require.NoError(t, err)
//	assert.Equal(t, fileData, b)
//}
//
//func TestSharedRepo_ReadFile_FileNotFound(t *testing.T) {
//	repo := initRepo(t, plumbing.NewBranchReferenceName(branchName))
//	sharedRepo, err := gitrepo.NewSingleTreeRepo(repo)
//	require.NoError(t, err)
//	_, err = sharedRepo.ReadFile(branchName, "not-"+fileName)
//	assert.True(t, os.IsNotExist(err))
//}

func initSingleTree(t *testing.T, refName plumbing.ReferenceName) *gitrepo.SingleTreeRepo {
	repo := initRepo(t, refName)
	sharedRepo, err := gitrepo.NewSingleTreeRepo(repo)
	require.NoError(t, err)
	return sharedRepo
}

func initRepo(t *testing.T, refName plumbing.ReferenceName) *git.Repository {
	repo, err := git.PlainInit(t.TempDir(), false)
	require.NoError(t, err)

	// If a ref is provided, initialize with const fileName+fileData
	if refName != "" {
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

		if refName.IsBranch() {
			err = wt.Checkout(&git.CheckoutOptions{
				Branch: refName,
			})
			require.NoError(t, err)
		}
	}

	return repo
}
