package cmd_test

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-github/v32/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thepwagner/action-update/cmd"
)

func TestRun(t *testing.T) {
	_ = os.Setenv("GITHUB_EVENT_NAME", "issue_comment")
	eventPath := filepath.Join(t.TempDir(), "event.json")
	err := ioutil.WriteFile(eventPath, []byte(`{}`), 0600)
	require.NoError(t, err)
	_ = os.Setenv("GITHUB_EVENT_PATH", eventPath)

	ctx := context.Background()
	err = cmd.Run(ctx, cmd.HandlersByEventName{
		"issue_comment": func(_ context.Context, env *cmd.Environment, evt interface{}) error {
			assert.Equal(t, eventPath, env.GitHubEventPath)
			assert.IsType(t, &github.IssueCommentEvent{}, evt)
			return nil
		},
	})
	assert.NoError(t, err)
}
