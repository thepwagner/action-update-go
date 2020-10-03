package updater

import (
	"context"
	"regexp"

	"github.com/moby/buildkit/frontend/dockerfile/command"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/thepwagner/action-update-go/updater/docker"
	updater2 "github.com/thepwagner/action-update/updater"
)

func (u *Updater) Dependencies(_ context.Context) ([]updater2.Dependency, error) {
	return docker.ExtractDockerfileDependencies(u.root, u.extractDockerfile)
}

var ghRelease = regexp.MustCompile(`https://github\.com/([^/]+)/([^/]+)/releases/download/([^/]+)/`)

func (u *Updater) extractDockerfile(parsed *parser.Result) ([]updater2.Dependency, error) {
	vars := docker.NewInterpolation(parsed)

	deps := make([]updater2.Dependency, 0)
	for _, instruction := range parsed.AST.Children {
		// Ignore everything but RUN instructions
		if instruction.Value != command.Run {
			continue
		}

		// Best-effort interpolate, then extract GitHub release URLs from the resulting commands:
		cmdLine := vars.Interpolate(instruction.Next.Value)
		for _, ghReleaseMatch := range ghRelease.FindAllStringSubmatch(cmdLine, -1) {
			repo := ghReleaseMatch[1]
			name := ghReleaseMatch[2]
			vers := ghReleaseMatch[3]
			deps = append(deps, updater2.Dependency{
				Path:    formatGitHubRelease(repo, name),
				Version: vers,
			})
		}
	}
	return deps, nil
}
