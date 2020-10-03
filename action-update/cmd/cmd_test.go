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

func TestExecute(t *testing.T) {
	_ = os.Setenv("GITHUB_EVENT_NAME", "issue_comment")
	eventPath := filepath.Join(t.TempDir(), "event.json")
	err := ioutil.WriteFile(eventPath, []byte(`{}`), 0600)
	require.NoError(t, err)
	_ = os.Setenv("GITHUB_EVENT_PATH", eventPath)

	ctx := context.Background()
	err = cmd.Execute(ctx, &cmd.Config{}, cmd.HandlersByEventName{
		"issue_comment": func(_ context.Context, evt interface{}) error {
			assert.IsType(t, &github.IssueCommentEvent{}, evt)
			return nil
		},
	})
	assert.NoError(t, err)
}
