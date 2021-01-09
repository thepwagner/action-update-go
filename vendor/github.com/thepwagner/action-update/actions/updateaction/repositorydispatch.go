package updateaction

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/go-github/v33/github"
	"github.com/sirupsen/logrus"
	"github.com/thepwagner/action-update/repo"
	"github.com/thepwagner/action-update/updater"
)

const (
	RepoDispatchActionUpdate = "update-dependency"
)

func (h *handler) RepositoryDispatch(ctx context.Context, evt *github.RepositoryDispatchEvent) error {
	switch evt.GetAction() {
	case RepoDispatchActionUpdate:
		// TODO: is this the right action for this dependency? (e.g. if workflow has multiple action-update-*)
		return h.repoDispatchActionUpdate(ctx, evt)
	default:
		return h.UpdateAll(ctx)
	}
}

func (h *handler) repoDispatchActionUpdate(ctx context.Context, evt *github.RepositoryDispatchEvent) error {
	var payload RepoDispatchActionUpdatePayload
	if err := json.Unmarshal(evt.ClientPayload, &payload); err != nil {
		return fmt.Errorf("decoding payload: %w", err)
	}
	update := updater.Update{
		Path: payload.Path,
		Next: payload.Next,
	}

	baseBranch := evt.GetRepo().GetDefaultBranch()
	branchName := h.branchNamer.Format(baseBranch, update)

	logrus.WithFields(logrus.Fields{
		"path":           update.Path,
		"version":        update.Next,
		"branch":         branchName,
		"feedback_owner": payload.Feedback.Owner,
		"feedback_name":  payload.Feedback.Name,
		"feedback_issue": payload.Feedback.IssueNumber,
	}).Debug("applying update from repository")
	r, err := h.repo()
	if err != nil {
		return fmt.Errorf("getting Repo: %w", err)
	}
	repoUpdater, err := h.repoUpdater(r)
	if err != nil {
		return fmt.Errorf("getting RepoUpdater: %w", err)
	}
	if payload.Updater != "" && repoUpdater.Updater.Name() != payload.Updater {
		logrus.WithField("updater", payload.Updater).Info("skipping event for other updaters")
		return nil
	}

	var success bool
	if payload.Feedback.IssueNumber != 0 {
		defer func() {
			logrus.Info("sending feedback to provided issue")
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()

			// Search for the PR that was created:
			gh := repo.NewGitHubClient(h.cfg.GitHubToken)
			prList, _, err := gh.PullRequests.List(ctx, evt.GetRepo().GetOwner().GetLogin(), evt.GetRepo().GetName(), &github.PullRequestListOptions{
				Head: branchName,
			})
			if err != nil {
				logrus.WithError(err).Warn("error looking for pull request")
			}

			// TODO: should be JSON? this checkpoint requires human validation and that's OK
			var feedbackBody string
			if len(prList) == 0 {
				feedbackBody = fmt.Sprintf("%s - %s - %v", evt.GetRepo().GetFullName(), branchName, success)
			} else {
				feedbackBody = prList[0].GetHTMLURL()
			}

			_, _, err = gh.Issues.CreateComment(ctx, payload.Feedback.Owner, payload.Feedback.Name, payload.Feedback.IssueNumber, &github.IssueComment{
				Body: github.String(feedbackBody),
			})
			if err != nil {
				logrus.WithError(err).Warn("error reporting feedback")
			}
		}()
	}

	ug := updater.NewUpdateGroup("", update)
	if err := repoUpdater.Update(ctx, baseBranch, branchName, ug); err != nil {
		return err
	}
	success = true
	return nil
}

type RepoDispatchActionUpdatePayload struct {
	Updater  string                                  `json:"updater"`
	Path     string                                  `json:"path"`
	Next     string                                  `json:"next"`
	Feedback RepoDispatchActionUpdatePayloadFeedback `json:"feedback"`
}

type RepoDispatchActionUpdatePayloadFeedback struct {
	Owner       string `json:"owner"`
	Name        string `json:"name"`
	IssueNumber int    `json:"issue"`
}
