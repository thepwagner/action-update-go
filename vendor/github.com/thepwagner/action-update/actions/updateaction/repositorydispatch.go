package updateaction

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/go-github/v34/github"
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
			pullRequest, err := h.findPullRequest(ctx, gh, evt, branchName, update)
			if err != nil {
				logrus.WithError(err).Warn("error finding created pull request")
			}

			// TODO: should be JSON? this checkpoint requires human validation and that's OK
			var comment string
			if pullRequest == nil {
				comment = fmt.Sprintf("%s - %s - %v", evt.GetRepo().GetFullName(), branchName, success)
			} else {
				comment = pullRequest.GetHTMLURL()
			}

			_, _, err = gh.Issues.CreateComment(ctx, payload.Feedback.Owner, payload.Feedback.Name, payload.Feedback.IssueNumber, &github.IssueComment{
				Body: github.String(comment),
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

func (h *handler) findPullRequest(ctx context.Context, gh *github.Client, evt *github.RepositoryDispatchEvent, baseBranch string, update updater.Update) (*github.PullRequest, error) {
	// List PRs on the base branch:
	prList, _, err := gh.PullRequests.List(ctx, evt.GetRepo().GetOwner().GetLogin(), evt.GetRepo().GetName(), &github.PullRequestListOptions{
		Head: baseBranch,
	})
	if err != nil {
		return nil, fmt.Errorf("listing pull requests: %w", err)
	}

	// Decode signed update group embedded in the PR body, return the first match
	for _, pr := range prList {
		log := logrus.WithField("pr_number", pr.GetNumber())
		signed := repo.ExtractSignedUpdateGroup(pr.GetBody())
		if signed == nil {
			log.Debug("ignoring PR")
			continue
		}
		updates, err := updater.VerifySignedUpdateGroup(h.cfg.SigningKey(), *signed)
		if err != nil {
			return nil, fmt.Errorf("verifying signature: %w", err)
		} else if updates == nil {
			log.Debug("updates not found")
			continue
		}

		for _, u := range updates.Updates {
			if u.Path == update.Path && u.Next == update.Next {
				return pr, nil
			}
		}
	}
	return nil, nil
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
