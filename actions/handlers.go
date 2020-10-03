package actions

import "github.com/thepwagner/action-update-go/cmd"

var Handlers = cmd.HandlersByEventName{
	"issue_comment":     IssueComment,
	"pull_request":      PullRequest,
	"schedule":          Schedule,
	"workflow_dispatch": Schedule,
}
