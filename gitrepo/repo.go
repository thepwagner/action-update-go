package gitrepo

import (
	"fmt"
	"io/ioutil"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/sirupsen/logrus"
	"github.com/thepwagner/action-update-go/gomod"
)

const RemoteName = "origin"

// SingleTreeRepo is a Repo that synchronizes access to a single git working tree.
type SingleTreeRepo struct {
	repo   *git.Repository
	wt     *git.Worktree
	branch string
}

var _ gomod.Repo = (*SingleTreeRepo)(nil)

// NewSingleTreeRepo creates SingleTreeRepo.
func NewSingleTreeRepo(repo *git.Repository) (*SingleTreeRepo, error) {
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

	var branch string
	if head, err := repo.Head(); err == nil && head.Name().IsBranch() {
		branch = head.Name().Short()
	}

	return &SingleTreeRepo{
		repo:   repo,
		wt:     wt,
		branch: branch,
	}, nil
}

func (t *SingleTreeRepo) Branch() string {
	return t.branch
}

func (t *SingleTreeRepo) SetBranch(branch string) error {
	log := logrus.WithField("branch", branch)
	log.Debug("switching branch")
	refName := plumbing.NewBranchReferenceName(branch)
	ref, err := t.repo.Reference(refName, true)
	if err != nil {
		if err != plumbing.ErrReferenceNotFound {
			return fmt.Errorf("querying branch ref: %w", err)
		}
		log.Debug("not found locally, checking remote")

		// If the ref doesn't exist, what about on the default remote?
		remoteRef, err := t.repo.Reference(plumbing.NewRemoteReferenceName(RemoteName, branch), true)
		if err != nil {
			return fmt.Errorf("querying remote branch ref: %w", err)
		}

		// Branch found on remote, store as local branch for later consistency:
		ref = plumbing.NewHashReference(refName, remoteRef.Hash())
		if err := t.repo.Storer.SetReference(ref); err != nil {
			return fmt.Errorf("storing reference: %w", err)
		}
	}

	if err := t.setBranch(ref.Name()); err != nil {
		return fmt.Errorf("switching to branch:%w", err)
	}
	return nil
}

func (t *SingleTreeRepo) setBranch(refName plumbing.ReferenceName) error {
	err := t.wt.Checkout(&git.CheckoutOptions{Branch: refName})
	if err != nil {
		return fmt.Errorf("checking out branch: %w", err)
	}
	t.branch = refName.Short()
	return nil
}

func (t *SingleTreeRepo) NewBranch(baseBranch, branch string) error {
	log := logrus.WithFields(logrus.Fields{
		"base":   baseBranch,
		"branch": branch,
	})
	log.Debug("creating branch")

	// Map string to a ref:
	baseRef, err := t.repo.Reference(plumbing.NewBranchReferenceName(baseBranch), true)
	if err != nil {
		if err != plumbing.ErrReferenceNotFound {
			return fmt.Errorf("querying branch ref: %w", err)
		}
		log.Debug("not found locally, checking remote")
		remoteRef, err := t.repo.Reference(plumbing.NewRemoteReferenceName(RemoteName, branch), true)
		if err != nil {
			return fmt.Errorf("querying remote branch ref: %w", err)
		}
		baseRef = remoteRef
	}

	// Create branch from ref and configure with remote:
	branchRefName := plumbing.NewBranchReferenceName(branch)
	if err := t.repo.Storer.SetReference(plumbing.NewHashReference(branchRefName, baseRef.Hash())); err != nil {
		return fmt.Errorf("creating branch reference: %w", err)
	}
	err = t.repo.CreateBranch(&config.Branch{
		Name:   branch,
		Merge:  branchRefName,
		Remote: RemoteName,
	})
	if err != nil {
		return fmt.Errorf("creating branch: %w", err)
	}
	log.WithField("base_ref", baseRef.Name()).Debug("branch created")
	if err := t.setBranch(branchRefName); err != nil {
		return fmt.Errorf("switching to new branch:%w", err)
	}
	return nil
}

func (t *SingleTreeRepo) Root() string {
	return t.wt.Filesystem.Root()
}

// ReadFile switches branch then reads a file. Stays on the requested branch but don't count on this.
func (t *SingleTreeRepo) ReadFile(branch, path string) ([]byte, error) {
	// Verify the base branch exists
	branchRef, err := ensureBranchExists(t.repo, branch)
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
func (t *SingleTreeRepo) NewSandbox(baseBranch, targetBranch string) (gomod.Sandbox, error) {
	//return NewSharedSandbox(&t.mu, t.repo, baseBranch, targetBranch)
	return nil, nil
}

func ensureBranchExists(repo *git.Repository, branch string) (*plumbing.Reference, error) {
	refName := plumbing.NewBranchReferenceName(branch)
	ref, err := repo.Reference(refName, true)
	if err != nil {
		if err != plumbing.ErrReferenceNotFound {
			return nil, fmt.Errorf("querying branch ref: %w", err)
		}

		// If the ref doesn't exist, what about on the default remote?
		remoteRef, err := repo.Reference(plumbing.NewRemoteReferenceName(RemoteName, branch), true)
		if err != nil {
			return nil, fmt.Errorf("querying remote remote branch ref: %w", err)
		}

		// Found on remote, store as base branch for later consistency:
		ref = plumbing.NewHashReference(refName, remoteRef.Hash())
		if err := repo.Storer.SetReference(ref); err != nil {
			return nil, fmt.Errorf("storing reference: %w", err)
		}
	}
	return ref, nil
}
