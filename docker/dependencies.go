package docker

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/semver"
	"github.com/moby/buildkit/frontend/dockerfile/command"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/thepwagner/action-update-go/updater"
)

func (u *Updater) Dependencies(ctx context.Context) ([]updater.Dependency, error) {
	deps := make([]updater.Dependency, 0)

	err := filepath.Walk(u.root, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() || !strings.HasPrefix(filepath.Base(path), "Dockerfile") {
			return nil
		}

		fileDeps, err := u.extractDockerfile(path)
		if err != nil {
			return fmt.Errorf("parsing %q: %w", path, err)
		}
		deps = append(deps, fileDeps...)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walking filesystem: %w", err)
	}

	return deps, nil
}

func (u *Updater) extractDockerfile(path string) ([]updater.Dependency, error) {
	parsed, err := u.parseDockerfile(path)
	if err != nil {
		return nil, err
	}

	buildArgs := extractBuildArgs(parsed)

	deps := make([]updater.Dependency, 0)
	for _, instruction := range parsed.AST.Children {
		switch instruction.Value {
		case command.From:
			image := instruction.Next.Value
			imageSplit := strings.SplitN(image, ":", 2)
			if len(imageSplit) == 1 {
				deps = append(deps, updater.Dependency{Path: image, Version: "latest"})
				continue
			}

			if strings.Contains(imageSplit[1], "$") {
				// Version contains a variable, attempt interpolation:
				vers := imageSplit[1]
				for k, v := range buildArgs {
					vers = strings.ReplaceAll(vers, fmt.Sprintf("${%s}", k), v)
					vers = strings.ReplaceAll(vers, fmt.Sprintf("$%s", k), v)
				}

				if !strings.Contains(vers, "$") {
					deps = append(deps, updater.Dependency{Path: imageSplit[0], Version: vers})
				}
			} else if semver.IsValid(imageSplit[1]) {
				deps = append(deps, updater.Dependency{Path: imageSplit[0], Version: imageSplit[1]})
			} else if s := fmt.Sprintf("v%s", imageSplit[1]); semver.IsValid(s) {
				deps = append(deps, updater.Dependency{Path: imageSplit[0], Version: imageSplit[1]})
			}
		}
	}
	return deps, nil
}

func extractBuildArgs(parsed *parser.Result) map[string]string {
	buildArgs := map[string]string{}
	for _, instruction := range parsed.AST.Children {
		if instruction.Value == command.Arg {
			varSplit := strings.SplitN(instruction.Next.Value, "=", 2)
			if len(varSplit) == 2 {
				buildArgs[varSplit[0]] = varSplit[1]
			}
		}
	}
	return buildArgs
}

func (u *Updater) parseDockerfile(path string) (*parser.Result, error) {
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
