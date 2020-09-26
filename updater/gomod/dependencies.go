package gomod

import (
	"context"
	"fmt"
	"io/ioutil"
	"path/filepath"

	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/modfile"
	"github.com/thepwagner/action-update-go/updater"
)

func (u *Updater) Dependencies(_ context.Context) ([]updater.Dependency, error) {
	b, err := ioutil.ReadFile(filepath.Join(u.root, GoModFn))
	if err != nil {
		return nil, fmt.Errorf("opening go.mod: %w", err)
	}
	parsed, err := modfile.Parse(GoModFn, b, nil)
	if err != nil {
		return nil, fmt.Errorf("parsing go.mod: %w", err)
	}

	deps := make([]updater.Dependency, 0, len(parsed.Require))
	for _, req := range parsed.Require {
		deps = append(deps, updater.Dependency{
			Path:     req.Mod.Path,
			Version:  req.Mod.Version,
			Indirect: req.Indirect,
		})
	}
	return deps, nil
}
