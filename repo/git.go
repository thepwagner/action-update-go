package repo

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
	"github.com/thepwagner/action-update-go/updater"
)

const RemoteName = "origin"

// GitRepo is a Repo that synchronizes access to a single git working tree.
type GitRepo struct {
	repo          *git.Repository
	wt            *git.Worktree
	branch        string
	author        GitIdentity
	remotes       bool
	commitMessage commitMessageGen
}

var _ updater.Repo = (*GitRepo)(nil)

// GitIdentity performs commits.
type GitIdentity struct {
	Name  string
	Email string
}

var DefaultGitIdentity = GitIdentity{
	Name:  "actions-update-go",
	Email: "noreply@github.com",
}

// NewGitRepo creates GitRepo.
func NewGitRepo(repo *git.Repository) (*GitRepo, error) {
	wt, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("getting work tree: %w", err)
	}

	// We'll likely be switching branches, so defensively ensure the tree is initially clean:
	if status, err := wt.Status(); err != nil {
		return nil, fmt.Errorf("getting worktree status: %w", err)
	} else if !status.IsClean() {
		for fn := range status {
			logrus.WithField("fn", fn).Warn("unexpected file in tree")
		}
		return nil, fmt.Errorf("tree is dirty")
	}

	remotes, err := repo.Remotes()
	if err != nil {
		return nil, fmt.Errorf("listing remotes: %w", err)
	}

	var branch string
	if head, err := repo.Head(); err == nil && head.Name().IsBranch() {
		branch = head.Name().Short()
	}
	return &GitRepo{
		repo:          repo,
		wt:            wt,
		branch:        branch,
		remotes:       len(remotes) > 0,
		author:        DefaultGitIdentity,
		commitMessage: defaultCommitMessage,
	}, nil
}

func (t *GitRepo) Branch() string {
	return t.branch
}

func (t *GitRepo) SetBranch(branch string) error {
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
		return err
	}
	return nil
}

func (t *GitRepo) setBranch(refName plumbing.ReferenceName) error {
	err := t.wt.Checkout(&git.CheckoutOptions{
		Branch: refName,
		Force:  true,
	})
	if err != nil {
		return fmt.Errorf("checking out branch: %w", err)
	}
	t.branch = refName.Short()
	return nil
}

func (t *GitRepo) NewBranch(base, branch string) error {
	log := logrus.WithFields(logrus.Fields{
		"base":   base,
		"branch": branch,
	})
	log.Debug("creating branch")

	// Map string to a ref:
	baseRef, err := t.repo.Reference(plumbing.NewBranchReferenceName(base), true)
	if err != nil {
		if err != plumbing.ErrReferenceNotFound {
			return fmt.Errorf("querying branch ref: %w", err)
		}
		log.Debug("not found locally, checking remote")
		remoteRef, err := t.repo.Reference(plumbing.NewRemoteReferenceName(RemoteName, base), true)
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
		return err
	}
	return nil
}

func (t *GitRepo) Root() string {
	return t.wt.Filesystem.Root()
}

func (t *GitRepo) Fetch(ctx context.Context, branch string) error {
	refName := fmt.Sprintf("refs/heads/%s", branch)
	cmd := exec.CommandContext(ctx, "git", "fetch", RemoteName, refName)
	cmd.Dir = t.Root()
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("fetching: %w", err)
	}
	return nil
}

func (t *GitRepo) Push(ctx context.Context, update ...updater.Update) error {
	commitMessage := t.commitMessage(update...)
	if err := t.commit(commitMessage); err != nil {
		return err
	}
	if err := t.push(ctx); err != nil {
		return err
	}
	return nil
}

func (t *GitRepo) commit(message string) error {
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

func (t *GitRepo) push(ctx context.Context) error {
	if !t.remotes {
		return nil
	}

	// go-git supports Push, but not the [http "https://github.com/"] .gitconfig that actions/checkout uses for auth
	// we could extract from u.repo.Config().Raw, but who are we trying to impress?
	cmd := exec.CommandContext(ctx, "git", "push", "-f")
	cmd.Dir = t.Root()
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("pushing: %w", err)
	}
	logrus.Debug("pushed to remote")
	return nil
}
