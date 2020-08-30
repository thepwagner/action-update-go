package gomod

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/modfile"
	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/semver"
	"github.com/sirupsen/logrus"
)

type Updater struct {
	Tidy bool
}

func (u *Updater) ApplyUpdate(ctx context.Context, root string, update Update) error {
	if err := updateGoMod(root, update); err != nil {
		return err
	}
	if update.Major() {
		if err := updateSourceCode(root, update); err != nil {
			return err
		}
	}

	if err := u.updateGoSum(ctx, root); err != nil {
		return err
	}

	if hasVendor(root) {
		if err := updateVendor(ctx, root); err != nil {
			return err
		}
	}
	return nil
}

func updateGoMod(root string, update Update) error {
	goModPath := filepath.Join(root, GoModFn)
	b, err := ioutil.ReadFile(goModPath)
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

	if err := ioutil.WriteFile(goModPath, updated, 0644); err != nil {
		return fmt.Errorf("writing updated go.mod: %w", err)
	}
	return nil
}

func updateSourceCode(root string, update Update) error {
	// replace foo.bar/v1 with foo.bar/v2 in imports:
	pattern, err := regexp.Compile(strings.ReplaceAll(update.Path, ".", "\\."))
	if err != nil {
		return err
	}

	pkgNext := path.Join(path.Dir(update.Path), semver.Major(update.Next))
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
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
		if err := updateSourceFile(path, pattern, pkgNext); err != nil {
			return err
		}
		return nil
	})
}

func updateSourceFile(path string, pattern *regexp.Regexp, replace string) error {
	f, err := os.OpenFile(path, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("reading source code file: %w", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			logrus.WithField("file_path", path).WithError(err).Warn("closing source code file")
		}
	}()

	var changed bool
	var importing bool
	in := bufio.NewScanner(f)
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
	logrus.WithField("file_path", path).Debug("updating go file")

	if _, err := f.Seek(0, 0); err != nil {
		return fmt.Errorf("resetting file offset: %w", err)
	}
	if _, err := f.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("writing updated source: %w", err)
	}
	return nil
}

func (u *Updater) updateGoSum(ctx context.Context, root string) error {
	// Shell out to the Go SDK for this, so the user has more control over generation:
	if err := rootGoCmd(ctx, root, "get", "-d", "-v"); err != nil {
		return fmt.Errorf("updating go.sum: %w", err)
	}

	if u.Tidy {
		if err := rootGoCmd(ctx, root, "mod", "tidy"); err != nil {
			return fmt.Errorf("tidying go.sum: %w", err)
		}
	}
	return nil
}

func rootGoCmd(ctx context.Context, root string, args ...string) error {
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = root
	if logrus.IsLevelEnabled(logrus.DebugLevel) {
		log := logrus.StandardLogger().WriterLevel(logrus.DebugLevel)
		defer func() { _ = log.Close() }()
		cmd.Stdout = log
		cmd.Stderr = log
		// echo command before output:
		_, _ = fmt.Fprintln(log, append([]string{"go"}, args...))
	}
	return cmd.Run()
}

func hasVendor(root string) bool {
	_, err := os.Stat(filepath.Join(root, VendorModulesFn))
	return err == nil
}

func updateVendor(ctx context.Context, root string) error {
	if err := rootGoCmd(ctx, root, "mod", "vendor"); err != nil {
		return fmt.Errorf("go vendoring: %w", err)
	}
	return nil
}
