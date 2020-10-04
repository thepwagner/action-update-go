package docker

import (
	"context"
	"fmt"
	"strings"

	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/semver"
	"github.com/moby/buildkit/frontend/dockerfile/command"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/thepwagner/action-update/updater"
)

func (u *Updater) Dependencies(_ context.Context) ([]updater.Dependency, error) {
	return ExtractDockerfileDependencies(u.root, u.extractDockerfile)
}

func (u *Updater) extractDockerfile(parsed *parser.Result) ([]updater.Dependency, error) {
	vars := NewInterpolation(parsed)

	deps := make([]updater.Dependency, 0)
	for _, instruction := range parsed.AST.Children {
		// Ignore everything but FROM instructions
		if instruction.Value != command.From {
			continue
		}

		// Parse the image name:
		image := instruction.Next.Value
		imageSplit := strings.SplitN(image, ":", 2)
		if len(imageSplit) == 1 {
			// No tag provided, default to ":latest"
			deps = append(deps, updater.Dependency{Path: image, Version: "latest"})
			continue
		}

		if strings.Contains(imageSplit[1], "$") {
			// Version contains a variable, attempt interpolation:
			vers := vars.Interpolate(imageSplit[1])
			if !strings.Contains(vers, "$") {
				deps = append(deps, updater.Dependency{Path: imageSplit[0], Version: vers})
			}
		} else if semver.IsValid(imageSplit[1]) {
			// Image tag is valid semver:
			deps = append(deps, updater.Dependency{Path: imageSplit[0], Version: imageSplit[1]})
		} else if s := fmt.Sprintf("v%s", imageSplit[1]); semver.IsValid(s) {
			// Image tag is close-enough to valid semver:
			deps = append(deps, updater.Dependency{Path: imageSplit[0], Version: imageSplit[1]})
		}
	}
	return deps, nil
}
