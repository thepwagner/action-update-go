package updateaction

import (
	"context"

	"github.com/sirupsen/logrus"
)

// UpdateAll tries to update all dependencies
func (h *handler) UpdateAll(ctx context.Context) error {
	// Open git repo, prepare updater:
	repo, err := h.repo()
	if err != nil {
		return err
	}
	repoUpdater, err := h.repoUpdater(repo)
	if err != nil {
		return err
	}

	// Capture initial branch, and revert when done:
	initialBranch := repo.Branch()
	defer func() {
		if err := repo.SetBranch(initialBranch); err != nil {
			logrus.WithError(err).Warn("error reverting to initial branch")
		}
	}()

	// If branches were provided as input, target those:
	if branches := h.cfg.Branches(); len(branches) > 0 {
		return repoUpdater.UpdateAll(ctx, branches...)
	}

	// No branches as input, fallback to current branch:
	return repoUpdater.UpdateAll(ctx, initialBranch)
}
