package gitrepo

import (
	"fmt"
	"io/ioutil"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/sirupsen/logrus"
	"github.com/thepwagner/action-update-go/gomod"
)

const RemoteName = "origin"

// SharedRepo is a Repo that synchronizes access to a single git working tree.
type SharedRepo struct {
	repo *git.Repository
	wt   *git.Worktree
	base *plumbing.Reference
	mu   sync.Mutex
}

var _ gomod.Repo = (*SharedRepo)(nil)

// NewSharedRepo creates SharedRepo.
func NewSharedRepo(repo *git.Repository) (*SharedRepo, error) {
	wt, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("getting work tree: %w", err)
	}

	// We'll likely be switching branches, so defensively ensure the tree is initially clean:
	if status, err := wt.Status(); err != nil {
		return nil, fmt.Errorf("getting worktree status: %w", err)
	} else if !status.IsClean() {
		return nil, fmt.Errorf("tree is dirty")
	}

	return &SharedRepo{
		repo: repo,
		wt:   wt,
	}, nil
}

// ReadFile switches branch then reads a file. Stays on the requested branch but don't count on this.
func (t *SharedRepo) ReadFile(branch, path string) ([]byte, error) {
	t.mu.Lock()
	defer t.mu.Unlock()

	// Verify the base branch exists
	branchRef, err := t.ensureBranchExists(branch)
	if err != nil {
		return nil, err
	}

	// Switch worktree to branch:
	if err := t.wt.Checkout(&git.CheckoutOptions{
		Hash:  branchRef.Hash(),
		Force: true,
	}); err != nil {
		return nil, fmt.Errorf("switching to branch to read file: %w", err)
	}
	logrus.WithFields(logrus.Fields{
		"branch": branch,
		"commit": branchRef.Hash(),
	}).Debug("switched shared work tree")

	// Read file from worktree:
	return readWorktreeFile(t.wt, path)
}

func readWorktreeFile(wt *git.Worktree, path string) ([]byte, error) {
	f, err := wt.Filesystem.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("reading from work tree %q: %w", path, err)
	}
	return b, nil
}

// NewSandbox creates a new sandbox branch for performing an update.
func (t *SharedRepo) NewSandbox(baseBranch, targetBranch string) (gomod.Sandbox, error) {
	t.mu.Lock()
	// passed to the returned SharedSandbox, which will Unlock() in Close()

	logrus.WithFields(logrus.Fields{
		"base":   baseBranch,
		"target": targetBranch,
	}).Debug("creating sandbox")
	baseBranchRef, err := t.ensureBranchExists(baseBranch)
	if err != nil {
		return nil, err
	}

	targetBranchRefName := plumbing.NewBranchReferenceName(targetBranch)
	err = t.wt.Checkout(&git.CheckoutOptions{
		Branch: targetBranchRefName,
		Hash:   baseBranchRef.Hash(),
		Create: true,
		Force:  true,
	})
	if err != nil {
		return nil, fmt.Errorf("checking out target branch: %w", err)
	}
	err = t.repo.CreateBranch(&config.Branch{
		Name:   targetBranch,
		Merge:  targetBranchRefName,
		Remote: "origin",
	})
	if err != nil {
		return nil, fmt.Errorf("creating target branch: %w", err)
	}
	return &SharedSandbox{
		lock: &t.mu,
		wt:   t.wt,
	}, nil
}

func (t *SharedRepo) ensureBranchExists(branch string) (*plumbing.Reference, error) {
	refName := plumbing.NewBranchReferenceName(branch)
	ref, err := t.repo.Reference(refName, true)
	if err != nil {
		if err != plumbing.ErrReferenceNotFound {
			return nil, fmt.Errorf("querying branch ref: %w", err)
		}

		// If the ref doesn't exist, what about on the default remote?
		originRef, err := t.repo.Reference(plumbing.NewRemoteReferenceName(RemoteName, branch), true)
		if err != nil {
			return nil, fmt.Errorf("querying remote remote branch ref: %w", err)
		}

		// Found on remote, store as base branch for later consistency:
		ref = plumbing.NewHashReference(refName, originRef.Hash())
		if err := t.repo.Storer.SetReference(ref); err != nil {
			return nil, fmt.Errorf("storing reference: %w", err)
		}
	}
	return ref, nil
}
