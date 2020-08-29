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

	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/modfile"
	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/semver"
	"github.com/sirupsen/logrus"
)

// Repo interfaces with Git repository
type Repo interface {
	ReadFile(branch, path string) ([]byte, error)
	NewSandbox(baseBranch, targetBranch string) (Sandbox, error)
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

func UpdateSandbox(ctx context.Context, sbx Sandbox, update ModuleUpdate, tidy bool) error {
	if err := updateGoMod(sbx, update); err != nil {
		return err
	}
	if update.Major() {
		if err := updateSourceCode(sbx, update); err != nil {
			return err
		}
	}

	if err := updateGoSum(ctx, sbx, tidy); err != nil {
		return err
	}

	if hasVendor(sbx) {
		if err := updateVendor(ctx, sbx); err != nil {
			return err
		}
	}
	return nil
}

func updateGoMod(sbx Sandbox, update ModuleUpdate) error {
	b, err := sbx.ReadFile(GoModFn)
	if err != nil {
		return fmt.Errorf("reading go.mod: %w", err)
	}
	goMod, err := modfile.Parse(GoModFn, b, nil)
	if err != nil {
		return fmt.Errorf("parsing go.mod: %w", err)
	}

	if update.Major() {
		// Replace foo.bar/v2 with foo.bar/v3:
		if err := goMod.DropRequire(update.Path); err != nil {
			return fmt.Errorf("dropping requirement: %w", err)
		}
		pkgNext := path.Join(path.Dir(update.Path), semver.Major(update.Next))
		if err := goMod.AddRequire(pkgNext, update.Next); err != nil {
			return fmt.Errorf("dropping requirement: %w", err)
		}
	} else {
		// Replace the version:
		if err := goMod.AddRequire(update.Path, update.Next); err != nil {
			return fmt.Errorf("adding requirement: %w", err)
		}
	}

	updated, err := goMod.Format()
	if err != nil {
		return fmt.Errorf("formatting go.mod: %w", err)
	}
	if logrus.IsLevelEnabled(logrus.DebugLevel) {
		out := logrus.StandardLogger().WriterLevel(logrus.DebugLevel)
		defer func() { _ = out.Close() }()
		_, _ = fmt.Fprintln(out, "-- go.mod --")
		_, _ = out.Write(updated)
		_, _ = fmt.Fprintln(out, "-- /go.mod --")
	}

	if err := sbx.WriteFile(GoModFn, updated); err != nil {
		return fmt.Errorf("writing updated go.mod: %w", err)
	}
	return nil
}

func updateSourceCode(sbx Sandbox, update ModuleUpdate) error {
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
