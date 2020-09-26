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
	parsed, err := u.parseGoMod()
	if err != nil {
		return nil, err
	}
	return extractDependencies(parsed), nil
}

func (u *Updater) parseGoMod() (*modfile.File, error) {
	b, err := ioutil.ReadFile(filepath.Join(u.root, GoModFn))
	if err != nil {
		return nil, fmt.Errorf("opening go.mod: %w", err)
	}
	parsed, err := modfile.Parse(GoModFn, b, nil)
	if err != nil {
		return nil, fmt.Errorf("parsing go.mod: %w", err)
	}
	return parsed, nil
}

func extractDependencies(parsed *modfile.File) []updater.Dependency {
	replacements := indexReplacements(parsed)

	deps := make([]updater.Dependency, 0, len(parsed.Require))
	for _, req := range parsed.Require {
		if replacement, ok := replacements[req.Mod.Path]; ok {
			replacement.Indirect = req.Indirect
			deps = append(deps, replacement)
			continue
		}

		deps = append(deps, updater.Dependency{
			Path:     req.Mod.Path,
			Version:  req.Mod.Version,
			Indirect: req.Indirect,
		})
	}
	return deps
}

func indexReplacements(parsed *modfile.File) map[string]updater.Dependency {
	replacements := map[string]updater.Dependency{}
	for _, replace := range parsed.Replace {
		replacements[replace.Old.Path] = updater.Dependency{
			Path:    replace.New.Path,
			Version: replace.New.Version,
		}
	}
	return replacements
}
