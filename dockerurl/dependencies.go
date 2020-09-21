package dockerurl

import (
	"context"
	"fmt"
	"regexp"

	"github.com/moby/buildkit/frontend/dockerfile/command"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/thepwagner/action-update-go/docker"
	"github.com/thepwagner/action-update-go/updater"
)

func (u *Updater) Dependencies(_ context.Context) ([]updater.Dependency, error) {
	return docker.WalkDockerfiles(u.root, u.extractDockerfile)
}

var ghRelease = regexp.MustCompile(`https://github.com/(\w+)/(\w+)/releases/download/([^/]+)/`)

func (u *Updater) extractDockerfile(parsed *parser.Result) ([]updater.Dependency, error) {
	vars := docker.NewInterpolation(parsed)

	deps := make([]updater.Dependency, 0)
	for _, instruction := range parsed.AST.Children {
		switch instruction.Value {
		case command.Run:
			cmdLine := vars.Interpolate(instruction.Next.Value)
			ghReleaseMatch := ghRelease.FindStringSubmatch(cmdLine)
			if len(ghReleaseMatch) > 0 {
				repo := ghReleaseMatch[1]
				name := ghReleaseMatch[2]
				vers := ghReleaseMatch[3]
				deps = append(deps, updater.Dependency{
					Path:    fmt.Sprintf("https://github.com/%s/%s/releases", repo, name),
					Version: vers,
				})
			}
		}
	}
	return deps, nil
}
