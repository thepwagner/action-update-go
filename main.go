package main

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/thepwagner/action-update-go/actions"
	"github.com/thepwagner/action-update-go/cmd"
)

var handlers = cmd.HandlersByEventName{
	"issue_comment":     actions.IssueComment,
	"pull_request":      actions.PullRequest,
	"schedule":          actions.Schedule,
	"workflow_dispatch": actions.Schedule,
}

func main() {
	ctx := context.Background()
	if err := cmd.Run(ctx, handlers); err != nil {
		logrus.WithError(err).Fatal("failed")
	}
}
