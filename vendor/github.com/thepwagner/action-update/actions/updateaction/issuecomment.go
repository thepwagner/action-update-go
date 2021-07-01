package updateaction

import (
	"context"

	"github.com/google/go-github/v36/github"
	"github.com/sirupsen/logrus"
)

// IssueComment is for debugging.
func IssueComment(_ context.Context, evt *github.IssueCommentEvent) error {
	body := evt.GetComment().GetBody()
	logrus.WithField("body", body).Info("issue comment")
	return nil
}
