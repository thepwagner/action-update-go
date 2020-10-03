package cmd_test

import (
	"bytes"
	"context"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thepwagner/action-update/cmd"
)

func TestCommandExecute_Debug(t *testing.T) {
	buf := logBuffer(t, logrus.DebugLevel)
	err := cmd.CommandExecute(context.Background(), "/", "sh", "-c", "echo foobar")
	require.NoError(t, err)
	assert.Contains(t, buf.String(), "foobar")
}

func TestCommandExecute_Warn(t *testing.T) {
	buf := logBuffer(t, logrus.WarnLevel)
	err := cmd.CommandExecute(context.Background(), "/", "sh", "-c", "echo foobar")
	require.NoError(t, err)
	assert.NotContains(t, buf.String(), "foobar")
}

func TestCommandExecute_Error(t *testing.T) {
	buf := logBuffer(t, logrus.WarnLevel)
	err := cmd.CommandExecute(context.Background(), "/", "sh", "-c", "echo foobar; false")
	assert.Error(t, err)
	assert.Contains(t, buf.String(), "foobar")
}

func logBuffer(t *testing.T, lvl logrus.Level) *bytes.Buffer {
	var buf bytes.Buffer
	logrus.SetOutput(&buf)
	ogLvl := logrus.GetLevel()
	t.Cleanup(func() { logrus.SetLevel(ogLvl) })
	logrus.SetLevel(lvl)
	return &buf
}
