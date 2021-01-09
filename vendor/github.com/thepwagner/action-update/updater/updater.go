package updater

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/sirupsen/logrus"
)

// RepoUpdater creates branches proposing all available updates for a Go module.
type RepoUpdater struct {
	repo        Repo
	Updater     Updater
	groups      Groups
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
	Push(context.Context, UpdateGroup) error
	// Fetch loads a remote ref without updating the working copy.
	Fetch(ctx context.Context, branch string) error
	// ExistingUpdates returns the state of recent updates
	ExistingUpdates(ctx context.Context, baseBranch string) (ExistingUpdates, error)
}

type Updater interface {
	Name() string
	Dependencies(context.Context) ([]Dependency, error)
	Check(ctx context.Context, dep Dependency, filter func(string) bool) (*Update, error)
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
		Updater:     updater,
		branchNamer: DefaultUpdateBranchNamer{},
	}
	for _, opt := range opts {
		opt(u)
	}
	return u
}

type RepoUpdaterOpt func(*RepoUpdater)

func WithGroups(groups ...*Group) RepoUpdaterOpt {
	return func(u *RepoUpdater) {
		u.groups = groups
	}
}

func WithBranchNamer(branchNamer UpdateBranchNamer) RepoUpdaterOpt {
	return func(u *RepoUpdater) {
		u.branchNamer = branchNamer
	}
}

// Update creates a single update branch included the Repo.
func (u *RepoUpdater) Update(ctx context.Context, baseBranch, branchName string, updates UpdateGroup) error {
	if err := u.repo.NewBranch(baseBranch, branchName); err != nil {
		return fmt.Errorf("switching to target branch: %w", err)
	}
	for _, update := range updates.Updates {
		if err := u.Updater.ApplyUpdate(ctx, update); err != nil {
			return fmt.Errorf("applying update: %w", err)
		}
	}

	if err := u.repo.Push(ctx, updates); err != nil {
		return fmt.Errorf("pushing update: %w", err)
	}
	return nil
}

// UpdateAll creates updates from a base branch included the Repo.
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
	deps, err := u.Updater.Dependencies(ctx)
	if err != nil {
		return fmt.Errorf("getting dependencies: %w", err)
	}

	// Load existing updates
	existing, err := u.repo.ExistingUpdates(ctx, branch)
	if err != nil {
		return fmt.Errorf("fetching existing updates: %w", err)
	}

	// Cluster dependencies into groups:
	groups, ungrouped := u.groups.GroupDependencies(deps)
	log.WithFields(logrus.Fields{
		"deps":      len(deps),
		"groups":    len(groups),
		"ungrouped": len(ungrouped),
	}).Info("parsed dependencies, checking for updates")

	updates := 0
	for groupName, groupDeps := range groups {
		groupLog := log.WithField("group", groupName)

		if cd := u.groups.ByName(groupName).CoolDownDuration(); cd > 0 {
			if latest := existing.LatestGroupUpdate(groupName); !latest.IsZero() {
				ageOfLatest := time.Since(latest)
				if ageOfLatest < cd {
					groupLog.WithFields(logrus.Fields{
						"cooldown": cd,
						"age":      ageOfLatest,
					}).Info("skipping group in cooldown")
					continue
				}
			}
		}

		groupLog.WithField("deps", len(groupDeps)).Debug("checking update group")
		groupUpdates, err := u.groupedUpdate(ctx, log, branch, groupName, groupDeps)
		if err != nil {
			groupLog.WithError(err).Error("error processing update group")
			continue
		}
		groupLog.WithField("updates", groupUpdates).Debug("checked update group")
		updates += groupUpdates
	}

	for _, dep := range ungrouped {
		ok, err := u.singleUpdate(ctx, log, branch, dep)
		if err != nil {
			logrus.WithField("path", dep.Path).WithError(err).Error("error processing update")
			continue
		}
		if ok {
			updates++
		}
	}

	log.WithFields(logrus.Fields{
		"deps":    len(deps),
		"updates": updates,
	}).Info("checked for updates")

	return nil
}

func (u *RepoUpdater) checkForUpdate(ctx context.Context, log logrus.FieldLogger, dep Dependency, filter func(string) bool) *Update {
	update, err := u.Updater.Check(ctx, dep, filter)
	if err != nil {
		log.WithError(err).Warn("error checking for updates")
		return nil
	}
	return update
}

func (u *RepoUpdater) singleUpdate(ctx context.Context, log logrus.FieldLogger, baseBranch string, dep Dependency) (bool, error) {
	depLog := log.WithField("path", dep.Path)
	update := u.checkForUpdate(ctx, depLog, dep, nil)
	if update == nil {
		return false, nil
	}

	// There is an update to apply
	updateLog := logrus.WithFields(logrus.Fields{
		"path":     update.Path,
		"previous": update.Previous,
		"next":     update.Next,
	})
	updateLog.Info("attempting update...")
	branch := u.branchNamer.Format(baseBranch, *update)
	if err := u.repo.NewBranch(baseBranch, branch); err != nil {
		return false, fmt.Errorf("switching to target branch: %w", err)
	}
	if err := u.Updater.ApplyUpdate(ctx, *update); err != nil {
		return false, fmt.Errorf("applying batched update: %w", err)
	}

	ug := NewUpdateGroup("", *update)
	if err := u.repo.Push(ctx, ug); err != nil {
		return false, fmt.Errorf("pushing update: %w", err)
	}
	updateLog.Info("update complete")

	return true, nil
}

func (u *RepoUpdater) groupedUpdate(ctx context.Context, log logrus.FieldLogger, baseBranch, groupName string, deps []Dependency) (int, error) {
	group := u.groups.ByName(groupName)

	// Iterate dependencies, collecting updates:
	var updates []Update
	for _, dep := range deps {
		// Is an update available for this dependency?
		depLog := log.WithField("path", dep.Path)
		update := u.checkForUpdate(ctx, depLog, dep, group.InRange)
		if update == nil {
			continue
		}
		// There is an update to apply
		updateLog := depLog.WithField("next_version", update.Next)
		updateLog.Debug("update available")
		updates = append(updates, *update)
	}
	if len(updates) == 0 {
		return 0, nil
	}

	// Updates are available, prepare branch:
	branch := u.branchNamer.FormatBatch(baseBranch, groupName)
	if err := u.repo.NewBranch(baseBranch, branch); err != nil {
		return 0, fmt.Errorf("switching to target branch: %w", err)
	}

	if err := u.updateScript(ctx, "pre", group.PreScript); err != nil {
		return 0, fmt.Errorf("executing pre-update script: %w", err)
	}

	for _, update := range updates {
		if err := u.Updater.ApplyUpdate(ctx, update); err != nil {
			return 0, fmt.Errorf("applying batched update: %w", err)
		}
	}

	if err := u.updateScript(ctx, "post", group.PostScript); err != nil {
		return 0, fmt.Errorf("executing pre-update script: %w", err)
	}

	ug := NewUpdateGroup(groupName, updates...)
	if err := u.repo.Push(ctx, ug); err != nil {
		return 0, fmt.Errorf("pushing update: %w", err)
	}
	return len(updates), nil
}

func (u *RepoUpdater) updateScript(ctx context.Context, label, script string) error {
	if script == "" {
		return nil
	}
	cmd := exec.CommandContext(ctx, "/bin/sh", "-c", script)
	cmd.Dir = u.repo.Root()
	out := os.Stdout
	_, _ = fmt.Fprintf(out, "--- start %s update script ---\n", label)
	cmd.Stdout = out
	cmd.Stderr = out
	if err := cmd.Run(); err != nil {
		return err
	}
	_, _ = fmt.Fprintf(out, "--- end %s update script ---\n", label)
	return nil
}
