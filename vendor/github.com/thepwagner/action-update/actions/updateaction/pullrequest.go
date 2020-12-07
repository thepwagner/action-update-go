package updateaction

import (
	"context"
	"fmt"

	"github.com/google/go-github/v33/github"
	"github.com/sirupsen/logrus"
	"github.com/thepwagner/action-update/repo"
	"github.com/thepwagner/action-update/updater"
)

func (h *handler) PullRequest(ctx context.Context, evt *github.PullRequestEvent) error {
	switch evt.GetAction() {
	case "reopened":
		return h.prReopened(ctx, evt)
	case "assigned", "unassigned", "review_requested", "review_request_removed", "labeled", "unlabeled",
		"opened", "edited", "closed", "ready_for_review", "locked", "unlocked":
		// pass
	default:
		logrus.WithField("action", evt.GetAction()).Warn("unexpected action")
	}
	return nil
}
func (h *handler) prReopened(ctx context.Context, evt *github.PullRequestEvent) error {
	pr := evt.GetPullRequest()
	base := pr.GetBase().GetRef()
	head := pr.GetHead().GetRef()
	log := logrus.WithFields(logrus.Fields{
		"base":   base,
		"head":   head,
		"number": pr.GetNumber(),
	})

	signed := repo.ExtractSignedUpdateGroup(pr.GetBody())
	if signed == nil {
		log.Info("ignoring PR")
		return nil
	}
	updates, err := updater.VerifySignedUpdateGroup(h.cfg.SigningKey(), *signed)
	if err != nil {
		return err
	} else if updates == nil {
		return nil
	}
	log.WithField("updates", len(updates.Updates)).Debug("validated update PR")

	r, err := h.repo()
	if err != nil {
		return err
	}
	repoUpdater, err := h.repoUpdater(r)
	if err != nil {
		return err
	}

	// Since actions/checkout will default to only the PR head ref, fetch the base ref before recreating:
	if err := r.Fetch(ctx, base); err != nil {
		return fmt.Errorf("fetching base: %w", err)
	}

	if err := repoUpdater.Update(ctx, base, head, *updates); err != nil {
		return fmt.Errorf("performing update: %w", err)
	}

	return nil
}
