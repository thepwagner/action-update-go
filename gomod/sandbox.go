package gomod

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/semver"
	"github.com/sirupsen/logrus"
)

// Repo interfaces with Git repository
type Repo interface {
	ReadFile(branch, path string) ([]byte, error)
	NewSandbox(baseBranch, targetBranch string) (Sandbox, error)

	// Root returns the working tree root.
	Root() string
	// Branch returns the current branch.
	Branch() string
	// SetBranch changes to an existing branch.
	SetBranch(branch string) error
	// NewBranch creates and changes to a new branch.
	NewBranch(baseBranch, branch string) error
}

// Sandbox is a filesystem containing full source code for an updatable go module
type Sandbox interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte) error
	// Cmd executes a command in the sandbox
	Cmd(ctx context.Context, cmd string, args ...string) error
	Walk(filepath.WalkFunc) error
	Commit(ctx context.Context, message string, author GitIdentity) error
	// Close frees any resources allocated to the sandbox
	Close() error
}

type GitIdentity struct {
	Name  string
	Email string
}

func updateSourceCode(sbx Sandbox, update Update) error {
	// replace foo.bar/v1 with foo.bar/v2 in imports:
	pattern, err := regexp.Compile(strings.ReplaceAll(update.Path, ".", "\\."))
	if err != nil {
		return err
	}

	pkgNext := path.Join(path.Dir(update.Path), semver.Major(update.Next))
	return sbx.Walk(func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logrus.WithError(err).WithField("path", path).Warn("error accessing path")
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}
		if err := updateSourceFile(sbx, path, pattern, pkgNext); err != nil {
			return err
		}
		return nil
	})
}

func updateSourceFile(sbx Sandbox, path string, pattern *regexp.Regexp, replace string) error {
	b, err := sbx.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading source code file: %w", err)
	}

	var changed bool
	var importing bool
	in := bufio.NewScanner(bytes.NewReader(b))
	var buf bytes.Buffer
	for in.Scan() {
		line := in.Text()
		if line == "import (" {
			importing = true
		} else if line == ")" && importing {
			importing = false
		}

		if importing || strings.HasPrefix(line, "import") {
			replaced := pattern.ReplaceAllString(line, replace)
			changed = changed || replaced != line
			line = replaced
		}

		_, _ = fmt.Fprintln(&buf, line)
	}
	if err := in.Err(); err != nil {
		return err
	}

	if !changed {
		return nil
	}

	if err := sbx.WriteFile(path, buf.Bytes()); err != nil {
		return fmt.Errorf("writing updated source: %w", err)
	}
	return nil
}

func updateGoSum(ctx context.Context, sbx Sandbox, tidy bool) error {
	// Shell out to the Go SDK for this, so the user has more control over generation:
	if err := sbx.Cmd(ctx, "go", "get", "-d", "-v"); err != nil {
		return fmt.Errorf("updating go.sum: %w", err)
	}

	if tidy {
		if err := sbx.Cmd(ctx, "go", "mod", "tidy"); err != nil {
			return fmt.Errorf("tidying go.sum: %w", err)
		}
	}
	return nil
}

func hasVendor(sbx Sandbox) bool {
	_, err := sbx.ReadFile(VendorModulesFn)
	return err == nil
}

func updateVendor(ctx context.Context, sbx Sandbox) error {
	if err := sbx.Cmd(ctx, "go", "mod", "vendor"); err != nil {
		return fmt.Errorf("go vendoring: %w", err)
	}
	return nil
}
