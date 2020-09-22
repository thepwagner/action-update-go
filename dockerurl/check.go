package dockerurl

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/semver"
	"github.com/google/go-github/v32/github"
	"github.com/thepwagner/action-update-go/updater"
)

func (u *Updater) Check(ctx context.Context, dependency updater.Dependency) (*updater.Update, error) {
	if strings.HasPrefix(dependency.Path, "github.com/") {
		return u.checkGitHubRelease(ctx, dependency)
	}
	return nil, fmt.Errorf("unknown dependency: %s", dependency.Path)
}

func (u *Updater) checkGitHubRelease(ctx context.Context, dependency updater.Dependency) (*updater.Update, error) {
	candidates, err := u.listGitHubReleases(ctx, dependency)
	if err != nil {
		return nil, err
	}

	if len(candidates) == 0 {
		return nil, nil
	}

	if semver.Compare(candidates[0], dependency.Version) > 0 {
		return &updater.Update{Path: dependency.Path, Previous: dependency.Version, Next: candidates[0]}, nil
	}
	return nil, nil
}

func (u *Updater) listGitHubReleases(ctx context.Context, dependency updater.Dependency) ([]string, error) {
	pathSplit := strings.SplitN(dependency.Path, "/", 4)
	repo := pathSplit[1]
	name := pathSplit[2]
	releases, _, err := u.github.ListReleases(ctx, repo, name, &github.ListOptions{PerPage: 100})
	if err != nil {
		return nil, fmt.Errorf("querying for releases: %w", err)
	}

	candidates := make([]string, 0, len(releases))
	for _, release := range releases {
		if release.GetDraft() || release.GetPrerelease() {
			continue
		}
		if !semver.IsValid(release.GetTagName()) {
			continue
		}
		// maybe filter alpha/beta?
		candidates = append(candidates, release.GetTagName())
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		return semver.Compare(candidates[i], candidates[j]) > 0
	})
	return candidates, nil
}
