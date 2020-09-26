package gomod

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
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
	"github.com/thepwagner/action-update-go/updater"
)

func (u *Updater) ApplyUpdate(ctx context.Context, up updater.Update) error {
	if err := u.updateGoMod(up); err != nil {
		return err
	}
	if MajorPkg(up) {
		if err := u.updateSourceCode(up); err != nil {
			return err
		}
	}

	if closer, err := u.ensureGoFileInRoot(); err != nil {
		return err
	} else if closer != nil {
		defer func() {
			if err := closer(); err != nil {
				logrus.WithError(err).Warn("cleaning up temp go file")
			}
		}()
	}

	if err := u.updateGoSum(ctx); err != nil {
		return err
	}

	if u.hasVendor() {
		if err := u.updateVendor(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (u *Updater) updateGoMod(update updater.Update) error {
	goModPath := filepath.Join(u.root, GoModFn)
	b, err := ioutil.ReadFile(goModPath)
	if err != nil {
		return fmt.Errorf("reading go.mod: %w", err)
	}
	goMod, err := modfile.Parse(GoModFn, b, nil)
	if err != nil {
		return fmt.Errorf("parsing go.mod: %w", err)
	}

	if err := patchParsedGoMod(goMod, update); err != nil {
		return err
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

func patchParsedGoMod(goMod *modfile.File, update updater.Update) error {
	// TODO: these can be combined (e.g. a major update via replacement)
	if MajorPkg(update) {
		if err := goMod.DropRequire(update.Path); err != nil {
			return fmt.Errorf("dropping requirement: %w", err)
		}
		pkgNext := path.Join(path.Dir(update.Path), semver.Major(update.Next))
		if err := goMod.AddRequire(pkgNext, update.Next); err != nil {
			return fmt.Errorf("dropping requirement: %w", err)
		}
		return nil
	}

	// Search for this path in the existing requirements:
	for _, req := range goMod.Require {
		if req.Mod.Path == update.Path {
			// Easiest case - replace the requirement
			if err := goMod.AddRequire(update.Path, update.Next); err != nil {
				return fmt.Errorf("adding requirement: %w", err)
			}
			return nil
		}
	}

	// Search for this path in the replacements:
	for _, rep := range goMod.Replace {
		if rep.New.Path == update.Path {
			if err := goMod.AddReplace(rep.Old.Path, rep.Old.Version, update.Path, update.Next); err != nil {
				return fmt.Errorf("dropping requirement: %w", err)
			}
			return nil
		}
	}

	return fmt.Errorf("could not update %q", update.Path)
}

func (u *Updater) updateSourceCode(up updater.Update) error {
	// replace foo.bar/v1 with foo.bar/v2 in imports:
	pattern, err := regexp.Compile(strings.ReplaceAll(up.Path, ".", "\\."))
	if err != nil {
		return err
	}

	pkgNext := path.Join(path.Dir(up.Path), semver.Major(up.Next))
	return filepath.Walk(u.root, func(path string, info os.FileInfo, err error) error {
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

func updateSourceFile(srcFile string, pattern *regexp.Regexp, replace string) error {
	f, err := os.OpenFile(srcFile, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("reading source code file: %w", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			logrus.WithField("file_path", srcFile).WithError(err).Warn("closing source code file")
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
	logrus.WithField("file_path", srcFile).Debug("updating go file")

	if _, err := f.Seek(0, 0); err != nil {
		return fmt.Errorf("resetting file offset: %w", err)
	}
	if _, err := f.Write(buf.Bytes()); err != nil {
		return fmt.Errorf("writing updated source: %w", err)
	}
	return nil
}

var fakeMainFile = []byte(`package main

func main() {}
`)

func (u *Updater) ensureGoFileInRoot() (func() error, error) {
	fileInfos, err := ioutil.ReadDir(u.root)
	if err != nil {
		return nil, fmt.Errorf("reading root dir: %w", err)
	}
	for _, fi := range fileInfos {
		if filepath.Ext(fi.Name()) == ".go" {
			return nil, nil
		}
	}

	// There is no .go file in the root, but having one makes `go get` easier.
	fakeMain := filepath.Join(u.root, "main.go")
	if err := ioutil.WriteFile(fakeMain, fakeMainFile, 0600); err != nil {
		return nil, fmt.Errorf("writing fake go file: %w", err)
	}
	return func() error {
		return os.Remove(fakeMain)
	}, nil
}

func (u *Updater) updateGoSum(ctx context.Context) error {
	// Shell out to the Go SDK for this, so the user has more control over generation:
	if err := u.rootGoCmd(ctx, "get", "-d", "-v"); err != nil {
		return fmt.Errorf("updating go.sum: %w", err)
	}

	if u.Tidy {
		if err := u.rootGoCmd(ctx, "mod", "tidy"); err != nil {
			return fmt.Errorf("tidying go.sum: %w", err)
		}
	}
	return nil
}

func (u *Updater) rootGoCmd(ctx context.Context, args ...string) error {
	cmd := exec.CommandContext(ctx, "go", args...)
	cmd.Dir = u.root

	// Capture output to buffer:
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
		_, _ = fmt.Fprintln(out, append([]string{"go"}, args...))
		_, _ = out.Write(buf.Bytes())
	}

	return err
}

func (u *Updater) hasVendor() bool {
	_, err := os.Stat(filepath.Join(u.root, VendorModulesFn))
	return err == nil
}

func (u *Updater) updateVendor(ctx context.Context) error {
	if err := u.rootGoCmd(ctx, "mod", "vendor"); err != nil {
		return fmt.Errorf("go vendoring: %w", err)
	}
	return nil
}
