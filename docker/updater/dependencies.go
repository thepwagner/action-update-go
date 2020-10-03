package updater

import (
	"context"
	"fmt"
	"strings"

	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/semver"
	"github.com/moby/buildkit/frontend/dockerfile/command"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	updater2 "github.com/thepwagner/action-update/updater"
)

func (u *Updater) Dependencies(_ context.Context) ([]updater2.Dependency, error) {
	return ExtractDockerfileDependencies(u.root, u.extractDockerfile)
}

var _ updater2.Updater = (*Updater)(nil)

func (u *Updater) extractDockerfile(parsed *parser.Result) ([]updater2.Dependency, error) {
	vars := NewInterpolation(parsed)

	deps := make([]updater2.Dependency, 0)
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
			deps = append(deps, updater2.Dependency{Path: image, Version: "latest"})
			continue
		}

		if strings.Contains(imageSplit[1], "$") {
			// Version contains a variable, attempt interpolation:
			vers := vars.Interpolate(imageSplit[1])
			if !strings.Contains(vers, "$") {
				deps = append(deps, updater2.Dependency{Path: imageSplit[0], Version: vers})
			}
		} else if semver.IsValid(imageSplit[1]) {
			// Image tag is valid semver:
			deps = append(deps, updater2.Dependency{Path: imageSplit[0], Version: imageSplit[1]})
		} else if s := fmt.Sprintf("v%s", imageSplit[1]); semver.IsValid(s) {
			// Image tag is close-enough to valid semver:
			deps = append(deps, updater2.Dependency{Path: imageSplit[0], Version: imageSplit[1]})
		}
	}
	return deps, nil
}
