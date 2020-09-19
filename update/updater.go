package update

import (
	"context"
	"fmt"

	"github.com/sirupsen/logrus"
)

// RepoUpdater creates branches proposing all available updates for a Go module.
type RepoUpdater struct {
	repo    Repo
	updater Updater
	Batch   bool
}

// Repo interfaces with an SCM repository, probably Git.
type Repo interface {
	// Root returns the working tree root.
	// This should minimally contain go.{mod,sum}. Vendoring or major updates require the full tree.
	Root() string
	// SetBranch changes to an existing branch.
	SetBranch(branch string) error
	// NewBranch creates and changes to a new branch.
	NewBranch(baseBranch string, update Update) error
	// Branch returns the current branch.
	Branch() string
	// Push snapshots the working tree after an update has been applied, and "publishes".
	// This is branch to commit. Publishing may mean push, create a PR, tweet the maintainer, whatever.
	Push(context.Context, Update) error
	// OpenUpdates returns any existing updates in the repo.
	Updates(context.Context) (UpdatesByBranch, error)
	// Parse matches a branch name that may be an update. Nil if not an update branch
	Parse(string) (baseBranch string, update *Update)
}

type Updater interface {
	Dependencies(context.Context) ([]Dependency, error)
	Check(context.Context, Dependency) (*Update, error)
	ApplyUpdate(context.Context, Update) error
}

// NewRepoUpdater creates RepoUpdater.
func NewRepoUpdater(repo Repo, updater Updater) *RepoUpdater {
	return &RepoUpdater{
		repo:    repo,
		updater: updater,
	}
}

// Update creates a single update branch in the Repo.
func (u *RepoUpdater) Update(ctx context.Context, baseBranch string, update Update) error {
	if err := u.repo.NewBranch(baseBranch, update); err != nil {
		return fmt.Errorf("switching to target branch: %w", err)
	}

	if err := u.updater.ApplyUpdate(ctx, update); err != nil {
		return fmt.Errorf("applying update: %w", err)
	}

	if err := u.repo.Push(ctx, update); err != nil {
		return fmt.Errorf("pushing update: %w", err)
	}

	return nil
}

// UpdateAll creates updates from a base branch in the Repo.
func (u *RepoUpdater) UpdateAll(ctx context.Context, branches ...string) error {
	updatesByBranch, err := u.repo.Updates(ctx)
	if err != nil {
		return fmt.Errorf("listing open updates: %w", err)
	}

	for _, branch := range branches {
		// Switch to base branch:
		if err := u.repo.SetBranch(branch); err != nil {
			return fmt.Errorf("switch to base branch: %w", err)
		}

		// List dependencies while on this branch:
		deps, err := u.updater.Dependencies(ctx)
		if err != nil {
			return fmt.Errorf("getting dependencies: %w", err)
		}
		log := logrus.WithField("branch", branch)
		log.WithField("deps", len(deps)).Info("parsed dependencies, checking for updates")

		// Iterate dependencies, collecting updates:
		existingUpdates := updatesByBranch[branch]
		var updates []Update
		for _, dep := range deps {
			// Is an update available for this dependency?
			depLog := log.WithField("path", dep.Path)
			update := u.checkForUpdate(ctx, depLog, existingUpdates, dep)
			if update == nil {
				continue
			}

			// There is an update to apply
			depLog = depLog.WithField("next_version", update.Next)
			updates = append(updates, *update)

			// When not batching, delegate to the standalone .Update() process
			if !u.Batch {
				if err := u.Update(ctx, branch, *update); err != nil {
					depLog.WithError(err).Warn("error applying update")
					continue
				}
			}
		}

		if u.Batch {
			if err := u.batchedUpdate(ctx, branch, updates); err != nil {
				return err
			}
		}

		log.WithFields(logrus.Fields{
			"deps":    len(deps),
			"updates": len(updates),
		}).Info("checked for updates")
	}
	return nil
}

func (u *RepoUpdater) checkForUpdate(ctx context.Context, log logrus.FieldLogger, existing Updates, dep Dependency) *Update {
	update, err := u.updater.Check(ctx, dep)
	if err != nil {
		log.WithError(err).Warn("error checking for updates")
		return nil
	}
	if update == nil {
		return nil
	}

	if existing := existing.Filter(*update); existing != "" {
		log.WithFields(logrus.Fields{
			"next_version":     update.Next,
			"existing_version": existing,
		}).Debug("existing version")
		return nil
	}
	return update
}

func (u *RepoUpdater) batchedUpdate(ctx context.Context, branch string, updates []Update) error {
	// XXX: better ideas here, this was quick
	batchUpdate := Update{
		Path: "github.com/thepwagner/action-update-go",
		Next: "BATCH",
	}

	if err := u.repo.NewBranch(branch, batchUpdate); err != nil {
		return fmt.Errorf("switching to target branch: %w", err)
	}

	for _, update := range updates {
		if err := u.updater.ApplyUpdate(ctx, update); err != nil {
			return fmt.Errorf("applying batched update: %w", err)
		}
	}

	if err := u.repo.Push(ctx, batchUpdate); err != nil {
		return fmt.Errorf("pushing update: %w", err)
	}
	return nil
}
