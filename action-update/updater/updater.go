package updater

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
)

// RepoUpdater creates branches proposing all available updates for a Go module.
type RepoUpdater struct {
	repo        Repo
	updater     Updater
	batchConfig map[string][]string
	branchNamer UpdateBranchNamer
}

// Repo interfaces with an SCM repository, probably Git.
type Repo interface {
	// Root returns the working tree root.
	// This should minimally contain go.{mod,sum}. Vendoring or major updates require the full tree.
	Root() string
	// SetBranch changes to an existing branch.
	SetBranch(branch string) error
	// NewBranch creates and changes to a new branch.
	NewBranch(base, branch string) error
	// Branch returns the current branch.
	Branch() string
	// Push snapshots the working tree after an update has been applied, and "publishes".
	// This is branch to commit. Publishing may mean push, create a PR, tweet the maintainer, whatever.
	Push(context.Context, ...Update) error
	// Fetch loads a remote ref without updating the working copy.
	Fetch(ctx context.Context, branch string) error
}

type Updater interface {
	Dependencies(context.Context) ([]Dependency, error)
	Check(context.Context, Dependency) (*Update, error)
	ApplyUpdate(context.Context, Update) error
}

// Factory provides UpdaterS for testing, masking any arguments other than the repo root.
type Factory interface {
	NewUpdater(root string) Updater
}

// NewRepoUpdater creates RepoUpdater.
func NewRepoUpdater(repo Repo, updater Updater, opts ...RepoUpdaterOpt) *RepoUpdater {
	u := &RepoUpdater{
		repo:        repo,
		updater:     updater,
		branchNamer: DefaultUpdateBranchNamer{},
	}
	for _, opt := range opts {
		opt(u)
	}
	return u
}

type RepoUpdaterOpt func(*RepoUpdater)

func WithBatches(batchConfig map[string][]string) RepoUpdaterOpt {
	return func(u *RepoUpdater) {
		u.batchConfig = batchConfig
	}
}

// Update creates a single update branch in the Repo.
func (u *RepoUpdater) Update(ctx context.Context, baseBranch, branchName string, updates ...Update) error {
	if err := u.repo.NewBranch(baseBranch, branchName); err != nil {
		return fmt.Errorf("switching to target branch: %w", err)
	}
	for _, update := range updates {
		if err := u.updater.ApplyUpdate(ctx, update); err != nil {
			return fmt.Errorf("applying update: %w", err)
		}
	}

	if err := u.repo.Push(ctx, updates...); err != nil {
		return fmt.Errorf("pushing update: %w", err)
	}
	return nil
}

// UpdateAll creates updates from a base branch in the Repo.
func (u *RepoUpdater) UpdateAll(ctx context.Context, branches ...string) error {
	multiBranch := len(branches) > 1
	for _, branch := range branches {
		var log logrus.FieldLogger
		if multiBranch {
			log = logrus.WithField("branch", branch)
		} else {
			log = logrus.StandardLogger()
		}
		if err := u.updateBranch(ctx, log, branch); err != nil {
			return err
		}
	}
	return nil
}

func (u *RepoUpdater) updateBranch(ctx context.Context, log logrus.FieldLogger, branch string) error {
	// Switch to base branch:
	if err := u.repo.SetBranch(branch); err != nil {
		return fmt.Errorf("switch to base branch: %w", err)
	}

	// List dependencies while on this branch:
	deps, err := u.updater.Dependencies(ctx)
	if err != nil {
		return fmt.Errorf("getting dependencies: %w", err)
	}
	batches := GroupDependencies(u.batchConfig, deps)
	log.WithFields(logrus.Fields{
		"deps":    len(deps),
		"batches": len(batches),
	}).Info("parsed dependencies, checking for updates")

	updates := 0
	for batchBranch, batchDeps := range batches {
		// Iterate dependencies, collecting updates:
		var batchUpdates []Update
		for _, dep := range batchDeps {
			// Is an update available for this dependency?
			depLog := log.WithField("path", dep.Path)
			update := u.checkForUpdate(ctx, depLog, dep)
			if update == nil {
				continue
			}
			// There is an update to apply
			depLog.WithField("next_version", update.Next).Debug("update available")
			batchUpdates = append(batchUpdates, *update)
		}

		if len(batchUpdates) == 0 {
			continue
		}

		if batchBranch != "" {
			if err := u.batchedUpdate(ctx, branch, batchBranch, batchUpdates); err != nil {
				return err
			}
		} else {
			if err := u.serialUpdates(ctx, branch, batchUpdates); err != nil {
				return err
			}
		}

		updates += len(batchUpdates)
	}
	log.WithFields(logrus.Fields{
		"deps":    len(deps),
		"updates": updates,
	}).Info("checked for updates")

	return nil
}

func (u *RepoUpdater) checkForUpdate(ctx context.Context, log logrus.FieldLogger, dep Dependency) *Update {
	update, err := u.updater.Check(ctx, dep)
	if err != nil {
		log.WithError(err).Warn("error checking for updates")
		return nil
	}
	if update == nil {
		return nil
	}

	return update
}

func (u *RepoUpdater) serialUpdates(ctx context.Context, base string, updates []Update) error {
	for _, update := range updates {
		updateLog := logrus.WithFields(logrus.Fields{
			"path":     update.Path,
			"previous": update.Previous,
			"next":     update.Next,
		})
		updateLog.Info("attempting update...")
		branch := u.branchNamer.Format(base, update)
		if err := u.repo.NewBranch(base, branch); err != nil {
			return fmt.Errorf("switching to target branch: %w", err)
		}
		if err := u.updater.ApplyUpdate(ctx, update); err != nil {
			return fmt.Errorf("applying batched update: %w", err)
		}
		if err := u.repo.Push(ctx, update); err != nil {
			return fmt.Errorf("pushing update: %w", err)
		}
		updateLog.Info("update complete")
	}
	return nil
}

func (u *RepoUpdater) batchedUpdate(ctx context.Context, base, batchName string, updates []Update) error {
	branch := u.branchNamer.FormatBatch(base, batchName)
	if err := u.repo.NewBranch(base, branch); err != nil {
		return fmt.Errorf("switching to target branch: %w", err)
	}

	for _, update := range updates {
		if err := u.updater.ApplyUpdate(ctx, update); err != nil {
			return fmt.Errorf("applying batched update: %w", err)
		}
	}

	if err := u.repo.Push(ctx, updates...); err != nil {
		return fmt.Errorf("pushing update: %w", err)
	}

	return nil
}
