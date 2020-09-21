package docker

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/thepwagner/action-update-go/updater"
)

func WalkDockerfiles(root string, extractor func(parsed *parser.Result) ([]updater.Dependency, error)) ([]updater.Dependency, error) {
	deps := make([]updater.Dependency, 0)

	err := filepath.Walk(root, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() || !strings.HasPrefix(filepath.Base(path), "Dockerfile") {
			return nil
		}

		parsed, err := parseDockerfile(path)
		if err != nil {
			return fmt.Errorf("parsing %q: %w", path, err)
		}
		fileDeps, err := extractor(parsed)
		if err != nil {
			return fmt.Errorf("extracting %q: %w", path, err)
		}
		deps = append(deps, fileDeps...)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walking filesystem: %w", err)
	}

	return deps, nil
}

func parseDockerfile(path string) (*parser.Result, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("opening dockerfile: %w", err)
	}
	defer f.Close()
	parsed, err := parser.Parse(f)
	if err != nil {
		return nil, fmt.Errorf("parsing dockerfile: %w", err)
	}
	return parsed, nil
}
