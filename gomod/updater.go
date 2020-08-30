package gomod

import (
	"context"
	"fmt"
	"io/ioutil"
	"path"
	"path/filepath"

	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/modfile"
	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/semver"
	"github.com/sirupsen/logrus"
)

type Updater struct {
	Tidy bool
}

func (u *Updater) ApplyUpdate(ctx context.Context, root string, update Update, tidy bool) error {
	if err := updateGoMod(root, update); err != nil {
		return err
	}
	//if update.Major() {
	//	if err := updateSourceCode(sbx, update); err != nil {
	//		return err
	//	}
	//}
	//
	//if err := updateGoSum(ctx, sbx, tidy); err != nil {
	//	return err
	//}
	//
	//if hasVendor(sbx) {
	//	if err := updateVendor(ctx, sbx); err != nil {
	//		return err
	//	}
	//}
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
