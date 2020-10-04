package actions_test

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thepwagner/action-update/actions"
)

func TestEnvironment_LogLevel(t *testing.T) {
	cases := map[string]logrus.Level{
		"":        logrus.InfoLevel,
		"invalid": logrus.InfoLevel,
		"warn":    logrus.WarnLevel,
	}

	for in, lvl := range cases {
		t.Run(fmt.Sprintf("parse %q", in), func(t *testing.T) {
			e := actions.Environment{InputLogLevel: in}
			assert.Equal(t, lvl, e.LogLevel())
		})
	}
}

func TestEnvironment_ParseEvent_Noop(t *testing.T) {
	e := actions.Environment{GitHubEventName: "schedule"}
	evt, err := e.ParseEvent()
	require.NoError(t, err)
	assert.Nil(t, evt)
}

func TestEnvironment_ParseEvent(t *testing.T) {
	e := actions.Environment{
		GitHubEventName: "issue_comment",
		GitHubEventPath: testIssueComment(t, "test"),
	}

	evt, err := e.ParseEvent()
	require.NoError(t, err)

	ic, ok := evt.(*github.IssueCommentEvent)
	if assert.True(t, ok) {
		assert.Equal(t, "test", ic.GetComment().GetBody())
	}
}

func testIssueComment(t *testing.T, body string) string {
	eventPath := filepath.Join(t.TempDir(), "event.json")
	err := ioutil.WriteFile(eventPath, []byte(fmt.Sprintf(`{"comment":{"body":%q}}`, body)), 0600)
	require.NoError(t, err)
	return eventPath
}
