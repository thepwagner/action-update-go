package main

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/thepwagner/action-update-go/handler"
	"github.com/thepwagner/jithub/api"
)

// FIXME: temp private dependency for testing
var _ *api.Handler = nil

var handlers = HandlersByEventName{
	"issue_comment":                  handler.IssueComment,
	"repository_vulnerability_alert": handler.RepositoryVulnerabilityAlert,
	"schedule":                       handler.Schedule,
	"workflow_dispatch":              handler.Schedule,
}

func main() {
	logrus.SetLevel(logrus.DebugLevel)

	ctx := context.Background()
	if err := Run(ctx, handlers); err != nil {
		logrus.WithError(err).Fatal("failed")
	}
}
