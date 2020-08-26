package main

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"

	"github.com/sirupsen/logrus"
	"github.com/thepwagner/action-update-go/cmd"
	"github.com/thepwagner/action-update-go/handler"
)

var handlers = cmd.HandlersByEventName{
	"issue_comment":                  handler.IssueComment,
	"repository_vulnerability_alert": handler.RepositoryVulnerabilityAlert,
	"schedule":                       handler.Schedule,
}

func main() {
	var buf bytes.Buffer
	checkSdk := exec.Command("go", "version")
	checkSdk.Stdout = &buf
	checkSdk.Stderr = &buf
	_ = checkSdk.Run()
	fmt.Println(buf.String())

	ctx := context.Background()
	if err := cmd.Run(ctx, handlers); err != nil {
		logrus.WithError(err).Fatal("failed")
	}
}
