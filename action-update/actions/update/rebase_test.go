package update

import (
	"context"
	"testing"

	"github.com/google/go-github/v32/github"
	"github.com/stretchr/testify/assert"
)

func TestPullRequest_UnhandledAction(t *testing.T) {
	err := prHandler(&github.PullRequestEvent{
		Action: github.String("unlocked"),
	})
	assert.NoError(t, err)
}

func TestPullRequest_Reopened_NoSignature(t *testing.T) {
	err := prHandler(&github.PullRequestEvent{
		Action: github.String("reopened"),
	})
	assert.NoError(t, err)
}

func TestPullRequest_Reopened_InvalidSignature(t *testing.T) {
	err := prHandler(&github.PullRequestEvent{
		Action: github.String("reopened"),
		PullRequest: &github.PullRequest{
			Body: github.String("<!--::action-update-go::{}-->"),
		},
	})
	assert.EqualError(t, err, "invalid signature")
}

func prHandler(evt *github.PullRequestEvent) error {
	ctx := context.Background()
	handler := NewHandlers(&Config{}, nil)["pull_request"]
	return handler(ctx, evt)
}
