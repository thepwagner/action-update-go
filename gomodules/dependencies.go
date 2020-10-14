package gomodules

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/sirupsen/logrus"
	"github.com/thepwagner/action-update/updater"
	"golang.org/x/mod/modfile"
)

func (u *Updater) Dependencies(_ context.Context) ([]updater.Dependency, error) {
	goModFiles, err := u.collectGoModFiles()
	if err != nil {
		return nil, err
	}
	logrus.WithField("gomods", len(goModFiles)).Debug("discovered go.mod files")

	deps, err := u.collectUniqueDependencies(goModFiles)
	if err != nil {
		return nil, err
	}

	return sortUniqueDependencies(deps)
}

func (u *Updater) collectGoModFiles() ([]string, error) {
	gomods, err := filepath.Glob(filepath.Join(u.root, "**", GoModFn))
	if err != nil {
		return nil, fmt.Errorf("collecting go.mod: %w", err)
	}
	rootGomod := filepath.Join(u.root, GoModFn)
	if _, err := os.Stat(rootGomod); err == nil {
		gomods = append(gomods, rootGomod)
	}
	return gomods, nil
}

func (u *Updater) collectUniqueDependencies(gomods []string) (map[string]updater.Dependency, error) {
	deps := map[string]updater.Dependency{}
	for _, gomod := range gomods {
		parsed, err := u.parseGoMod(gomod)
		if err != nil {
			return nil, err
		}

		for _, d := range extractDependencies(parsed) {
			if d.Version == "" {
				// Modules without versions are path replacements we can't affect:
				continue
			}
			depKey := fmt.Sprintf("%s-%s", d.Path, d.Version)
			deps[depKey] = d
		}
	}
	return deps, nil
}

func (u *Updater) parseGoMod(path string) (*modfile.File, error) {
	b, err := ioutil.ReadFile(path)
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

func sortUniqueDependencies(deps map[string]updater.Dependency) ([]updater.Dependency, error) {
	ret := make([]updater.Dependency, 0, len(deps))
	for _, d := range deps {
		ret = append(ret, d)
	}
	sort.Slice(ret, func(i, j int) bool {
		return ret[i].Path < ret[j].Path
	})
	return ret, nil
}
