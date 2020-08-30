package gitrepo

import (
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/sirupsen/logrus"
	"github.com/thepwagner/action-update-go/gomod"
)

const RemoteName = "origin"

type GitIdentity struct {
	Name  string
	Email string
}

// SingleTreeRepo is a Repo that synchronizes access to a single git working tree.
type SingleTreeRepo struct {
	repo   *git.Repository
	wt     *git.Worktree
	branch string
	author GitIdentity
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
		author: GitIdentity{
			Name:  "actions-update-go",
			Email: "noreply@github.com",
		},
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

func (t *SingleTreeRepo) Push(ctx context.Context, update gomod.Update) error {
	// TODO: dependency inject this?
	commitMessage := fmt.Sprintf("update %s to %s", update.Path, update.Next)
	if err := t.commit(commitMessage); err != nil {
		return err
	}
	if err := t.push(ctx); err != nil {
		return err
	}
	return nil
}

func (t *SingleTreeRepo) commit(message string) error {
	when := time.Now()
	if err := worktreeAddAll(t.wt); err != nil {
		return err
	}

	commit, err := t.wt.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  t.author.Name,
			Email: t.author.Email,
			When:  when,
		},
		Committer: &object.Signature{
			Name:  t.author.Name,
			Email: t.author.Email,
			When:  when,
		},
	})
	if err != nil {
		return fmt.Errorf("committing branch: %w", err)
	}
	logrus.WithField("commit", commit.String()).Info("committed update")
	return nil
}

func worktreeAddAll(wt *git.Worktree) error {
	// wt.AddGlob() is attractive, but does not respect .gitignore
	// .Status() respects .gitignore so add file by file:
	status, err := wt.Status()
	if err != nil {
		return fmt.Errorf("checking status for add: %w", err)
	}
	for fn := range status {
		if _, err := wt.Add(fn); err != nil {
			return fmt.Errorf("adding file %q: %w", fn, err)
		}
	}

	logrus.WithField("files", len(status)).Debug("added files to index")
	return nil
}

func (t *SingleTreeRepo) push(ctx context.Context) error {
	// go-git supports Push, but not the [http "https://github.com/"] .gitconfig directive that actions/checkout uses for auth
	// we could extract from u.repo.Config().Raw, but who are we trying to impress?
	cmd := exec.CommandContext(ctx, "git", "push", "-f")
	cmd.Dir = t.Root()
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pushing: %w", err)
	}
	return nil
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
