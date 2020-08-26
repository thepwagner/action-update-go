package handler

import (
	"context"
	"fmt"

	"github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"
)

func IssueComment(_ context.Context, evt interface{}) error {
	issueComment, ok := evt.(*github.IssueCommentEvent)
	if !ok {
		return fmt.Errorf("unexpected event")
	}
	body := issueComment.GetComment().GetBody()
	logrus.WithField("body", body).Info("issue comment")
	return nil
}

var _ Handler = IssueComment
