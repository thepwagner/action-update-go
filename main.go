package main

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/thepwagner/action-update-go/handler"
)

var handlers = HandlersByEventName{
	"issue_comment":                  handler.IssueComment,
	"pull_request":                   handler.PullRequest,
	"repository_vulnerability_alert": handler.RepositoryVulnerabilityAlert,
	"schedule":                       handler.Schedule,
	"workflow_dispatch":              handler.Schedule,
}

func main() {
	ctx := context.Background()
	if err := Run(ctx, handlers); err != nil {
		logrus.WithError(err).Fatal("failed")
	}
}
