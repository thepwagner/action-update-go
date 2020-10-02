package actions

import (
	"context"
	"testing"

	"github.com/google/go-github/v32/github"
	"github.com/stretchr/testify/assert"
	"github.com/thepwagner/action-update-go/cmd"
)

func TestPullRequest_UnhandledAction(t *testing.T) {
	ctx := context.Background()
	err := PullRequest(ctx, nil, &github.PullRequestEvent{
		Action: github.String("unlocked"),
	})
	assert.NoError(t, err)
}

func TestPullRequest_Reopened_NoSignature(t *testing.T) {
	ctx := context.Background()
	err := PullRequest(ctx, &cmd.Environment{}, &github.PullRequestEvent{
		Action: github.String("reopened"),
	})
	assert.NoError(t, err)
}

func TestPullRequest_Reopened_InvalidSignature(t *testing.T) {
	ctx := context.Background()
	err := PullRequest(ctx, &cmd.Environment{}, &github.PullRequestEvent{
		Action: github.String("reopened"),
		PullRequest: &github.PullRequest{
			Body: github.String("<!--::action-update-go::{}-->"),
		},
	})
	assert.EqualError(t, err, "invalid signature")
}
