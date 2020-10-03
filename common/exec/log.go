package exec

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
		_, _ = fmt.Fprintln(out, append([]string{name}, args...))
		_, _ = out.Write(buf.Bytes())
	}
	return err
}
