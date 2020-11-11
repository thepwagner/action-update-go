package updateaction

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"
	"github.com/thepwagner/action-update/repo"
)

func (h *handler) Release(ctx context.Context, evt *github.ReleaseEvent) error {
	if evt.GetAction() != "released" {
		logrus.WithField("action", evt.GetAction()).Info("ignoring release event")
		return nil
	}
	notifyRepos := h.cfg.ReleaseDispatchRepos()
	logrus.WithField("repos", len(notifyRepos)).Info("notifying repositories of release")
	if len(notifyRepos) == 0 {
		return nil
	}

	gh := repo.NewGitHubClient(h.cfg.GitHubToken)
	feedbackIssue, err := releaseFeedbackIssue(ctx, gh, evt, notifyRepos)
	if err != nil {
		return err
	}
	logrus.WithField("issue_number", feedbackIssue.Number).Debug("created feedback issue")

	dispatchOpts, err := releaseDispatchOptions(evt, feedbackIssue)
	if err != nil {
		return err
	}

	for _, notifyRepo := range notifyRepos {
		notifyRepoParts := strings.SplitN(notifyRepo, "/", 2)
		owner := notifyRepoParts[0]
		name := notifyRepoParts[1]
		logrus.WithFields(logrus.Fields{
			"owner": owner,
			"name":  name,
		}).Debug("dispatching release to repository")
		if _, _, err := gh.Repositories.Dispatch(ctx, owner, name, dispatchOpts); err != nil {
			logrus.WithError(err).Warn("error dispatching update")
		}
	}

	return nil
}

func releaseFeedbackIssue(ctx context.Context, gh *github.Client, evt *github.ReleaseEvent, repos []string) (*github.Issue, error) {
	ghRepo := evt.GetRepo()

	var body strings.Builder
	body.WriteString("Expecting feedback:\n\n")
	for _, r := range repos {
		_, _ = fmt.Fprintf(&body, "- [ ] %s\n", r)
	}

	issue, _, err := gh.Issues.Create(ctx, ghRepo.GetOwner().GetLogin(), ghRepo.GetName(), &github.IssueRequest{
		Title: github.String(fmt.Sprintf("Release feedback: %s", evt.GetRelease().GetTagName())),
		Body:  github.String(body.String()),
	})
	return issue, err
}

func releaseDispatchOptions(evt *github.ReleaseEvent, feedbackIssue *github.Issue) (github.DispatchRequestOptions, error) {
	payload, err := json.Marshal(&RepoDispatchActionUpdatePayload{
		Path: fmt.Sprintf("github.com/%s", evt.GetRepo().GetFullName()),
		Next: evt.GetRelease().GetTagName(),
		Feedback: RepoDispatchActionUpdatePayloadFeedback{
			Owner:       evt.GetRepo().GetOwner().GetLogin(),
			Name:        evt.GetRepo().GetName(),
			IssueNumber: feedbackIssue.GetNumber(),
		},
	})
	if err != nil {
		return github.DispatchRequestOptions{}, fmt.Errorf("serializing payload: %w", err)
	}
	clientPayload := json.RawMessage(payload)
	return github.DispatchRequestOptions{
		EventType:     RepoDispatchActionUpdate,
		ClientPayload: &clientPayload,
	}, nil
}
