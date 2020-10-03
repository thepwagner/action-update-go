package actions

import (
	"context"
	"fmt"

	"github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"
	"github.com/thepwagner/action-update-go/cmd"
	"github.com/thepwagner/action-update-go/repo"
	"github.com/thepwagner/action-update-go/updater"
)

func PullRequest(ctx context.Context, env *cmd.Environment, evt interface{}) error {
	pr, ok := evt.(*github.PullRequestEvent)
	if !ok {
		return fmt.Errorf("invalid event type: %T", evt)
	}

	switch pr.GetAction() {
	case "reopened":
		return prReopened(ctx, env, pr)
	case "assigned", "unassigned", "review_requested", "review_request_removed", "labeled", "unlabeled",
		"opened", "edited", "closed", "ready_for_review", "locked", "unlocked":
		// pass
	default:
		logrus.WithField("action", pr.GetAction()).Warn("unexpected action")
	}
	return nil
}

var _ cmd.Handler = PullRequest

func prReopened(ctx context.Context, env *cmd.Environment, evt *github.PullRequestEvent) error {
	pr := evt.GetPullRequest()
	base := pr.GetBase().GetRef()
	head := pr.GetHead().GetRef()
	log := logrus.WithFields(logrus.Fields{
		"base":   base,
		"head":   head,
		"number": pr.GetNumber(),
	})

	signed := repo.ExtractSignedUpdateDescriptor(pr.GetBody())
	if signed == nil {
		log.Info("ignoring PR")
		return nil
	}
	updates, err := updater.VerifySignedUpdateDescriptor(env.InputSigningKey, *signed)
	if err != nil {
		return err
	}
	log.WithField("updates", len(updates)).Debug("validated update PR")

	r, repoUpdater, err := getRepoUpdater(env)
	if err != nil {
		return err
	}

	if err := r.Fetch(ctx, base); err != nil {
		return fmt.Errorf("fetching base: %w", err)
	}

	if err := repoUpdater.Update(ctx, base, head, updates...); err != nil {
		return fmt.Errorf("performing update: %w", err)
	}

	return nil
}
