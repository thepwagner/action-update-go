package cmd

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os/exec"

	"github.com/sirupsen/logrus"
)

func CommandExecute(ctx context.Context, dir, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Dir = dir

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()

	var out io.WriteCloser
	if err != nil {
		out = logrus.StandardLogger().WriterLevel(logrus.ErrorLevel)
	} else if logrus.IsLevelEnabled(logrus.DebugLevel) {
		out = logrus.StandardLogger().WriterLevel(logrus.DebugLevel)
	}

	if out != nil {
		defer func() { _ = out.Close() }()
		// echo command before output:
		_, _ = fmt.Fprintf(out, "Command: %s", name)
		for _, a := range args {
			_, _ = fmt.Fprintf(out, " %q", a)
		}
		_, _ = fmt.Fprintln(out)
		_, _ = out.Write(buf.Bytes())
	}
	return err
}
