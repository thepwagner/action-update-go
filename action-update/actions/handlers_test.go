package actions_test

import (
	"context"
	"os"
	"testing"

	"github.com/google/go-github/v32/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thepwagner/action-update/actions"
)

func TestHandlers_Handle_NoHandler(t *testing.T) {
	h := &actions.Handlers{}

	err := h.Handle(context.Background(), &actions.Environment{
		GitHubEventName: "issue_comment",
	})
	require.NoError(t, err)
}

func TestHandlers_Handle(t *testing.T) {
	eventPath := testIssueComment(t, "sup")

	var called bool
	h := &actions.Handlers{
		IssueComment: func(_ context.Context, evt *github.IssueCommentEvent) error {
			called = true
			assert.Equal(t, "sup", evt.GetComment().GetBody())
			return nil
		},
	}

	err := h.Handle(context.Background(), &actions.Environment{
		GitHubEventName: "issue_comment",
		GitHubEventPath: eventPath,
	})
	require.NoError(t, err)
	assert.True(t, called)
}

func TestHandlers_ParseAndHandle_BadEnvironment(t *testing.T) {
	h := &actions.Handlers{}
	err := h.ParseAndHandle(context.Background(), nil)
	assert.EqualError(t, err, "env should embed actions.Environment")
}

func TestHandlers_ParseAndHandle(t *testing.T) {
	eventPath := testIssueComment(t, "sup")

	var called bool
	h := &actions.Handlers{
		IssueComment: func(_ context.Context, evt *github.IssueCommentEvent) error {
			called = true
			assert.Equal(t, "sup", evt.GetComment().GetBody())
			return nil
		},
	}

	_ = os.Setenv("GITHUB_EVENT_NAME", "issue_comment")
	_ = os.Setenv("GITHUB_EVENT_PATH", eventPath)
	err := h.ParseAndHandle(context.Background(), &actions.Environment{})
	require.NoError(t, err)
	assert.True(t, called)
}
