package handler

import (
	"context"
	"fmt"

	"github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"
	"github.com/thepwagner/action-update-go/cmd"
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

func prReopened(ctx context.Context, env *cmd.Environment, pr *github.PullRequestEvent) error {
	_, updater, err := getRepoUpdater(env)
	if err != nil {
		return err
	}

	prRef := pr.GetPullRequest().GetHead().GetRef()
	logrus.WithField("ref", prRef).Info("PR reopened, recreating update")

	base, update := updater.Parse(prRef)
	if update == nil {
		logrus.Info("not an update PR")
		return nil
	}

	if err := updater.Update(ctx, base, *update); err != nil {
		return fmt.Errorf("performing update: %w", err)
	}
	return nil
}
