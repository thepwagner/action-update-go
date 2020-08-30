package gomod

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/modfile"
	"github.com/sirupsen/logrus"
)

const (
	GoModFn         = "go.mod"
	VendorModulesFn = "vendor/modules.txt"
)

// RepoUpdater creates update branches for all potential updates to a Go module.
type RepoUpdater struct {
	repo       Repo
	branchName UpdateBranchNamer
	updater    *Updater
}

// Repo interfaces with (probably Git) repository
type Repo interface {
	// Root returns the working tree root.
	Root() string
	// SetBranch changes to an existing branch.
	SetBranch(branch string) error
	// NewBranch creates and changes to a new branch.
	NewBranch(baseBranch, branch string) error
	// Push snapshots and publishes (commit. then push, create PR, tweet, whatever).
	Push(ctx context.Context, update Update) error
}

func NewRepoUpdater(repo Repo) (*RepoUpdater, error) {
	u := &RepoUpdater{
		repo:       repo,
		branchName: DefaultUpdateBranchNamer,
		updater:    &Updater{Tidy: true},
	}
	return u, nil
}

func (u *RepoUpdater) UpdateAll(ctx context.Context, branch string) error {
	// Switch to base branch:
	if err := u.repo.SetBranch(branch); err != nil {
		return fmt.Errorf("switch to base branch: %w", err)
	}

	// Parse go.mod, to list updatable dependencies:
	goMod, err := u.parseGoMod()
	if err != nil {
		return fmt.Errorf("parsing go.mod: %w", err)
	}
	log := logrus.WithField("branch", branch)
	log.WithField("deps", len(goMod.Require)).Info("parsed go.mod, checking for updates")

	checker := &UpdateChecker{
		MajorVersions: true,
		RootDir:       u.repo.Root(),
	}

	// Iterate dependencies:
	for _, req := range goMod.Require {
		update, err := checker.CheckForModuleUpdates(ctx, req)
		if err != nil {
			log.WithError(err).Warn("error checking for updates")
			continue
		}
		if update == nil {
			continue
		}

		if err := u.update(ctx, branch, *update); err != nil {
			return fmt.Errorf("updating %q: %w", update.Path, err)
		}
	}
	return nil
}

func (u *RepoUpdater) parseGoMod() (*modfile.File, error) {
	b, err := ioutil.ReadFile(filepath.Join(u.repo.Root(), GoModFn))
	if err != nil {
		return nil, fmt.Errorf("opening go.mod: %w", err)
	}
	parsed, err := modfile.Parse(GoModFn, b, nil)
	if err != nil {
		return nil, fmt.Errorf("parsing go.mod: %w", err)
	}
	return parsed, nil
}

func (u *RepoUpdater) update(ctx context.Context, baseBranch string, update Update) error {
	targetBranch := u.branchName(baseBranch, update)
	if err := u.repo.NewBranch(baseBranch, targetBranch); err != nil {
		return fmt.Errorf("switching to target branch: %w", err)
	}

	if err := u.updater.ApplyUpdate(ctx, u.repo.Root(), update); err != nil {
		return fmt.Errorf("applying update: %w", err)
	}

	if err := u.repo.Push(ctx, update); err != nil {
		return fmt.Errorf("pushing update: %w", err)
	}

	return nil
}
