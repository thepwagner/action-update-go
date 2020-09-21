package docker

import (
	"context"
	"fmt"
	"strings"

	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/semver"
	"github.com/moby/buildkit/frontend/dockerfile/command"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/thepwagner/action-update-go/updater"
)

func (u *Updater) Dependencies(_ context.Context) ([]updater.Dependency, error) {
	return WalkDockerfiles(u.root, u.extractDockerfile)
}

var _ updater.Updater = (*Updater)(nil)

func (u *Updater) extractDockerfile(parsed *parser.Result) ([]updater.Dependency, error) {
	vars := NewInterpolation(parsed)

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
				vers := vars.Interpolate(imageSplit[1])
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
