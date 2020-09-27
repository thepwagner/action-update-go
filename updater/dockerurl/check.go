package dockerurl

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/semver"
	"github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"
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

	latest := candidates[0]
	log := logrus.WithFields(logrus.Fields{
		"path":            dependency.Path,
		"latest_version":  latest,
		"current_version": dependency.Version,
	})
	if semver.Compare(latest, dependency.Version) > 0 {
		log.Info("update available")
		return &updater.Update{Path: dependency.Path, Previous: dependency.Version, Next: latest}, nil
	}
	log.Debug("no update available")
	return nil, nil
}

func (u *Updater) listGitHubReleases(ctx context.Context, dependency updater.Dependency) ([]string, error) {
	owner, name := parseGitHubRelease(dependency.Path)
	releases, _, err := u.ghRepos.ListReleases(ctx, owner, name, &github.ListOptions{PerPage: 100})
	if err != nil {
		return nil, fmt.Errorf("querying for releases: %w", err)
	}
	log := logrus.WithFields(logrus.Fields{
		"owner": owner,
		"repo":  name,
	})
	log.WithField("releases", len(releases)).Debug("fetched releases")

	candidates := make([]string, 0, len(releases))
	prereleases := make(map[string]struct{}, len(releases))
	for _, release := range releases {
		if release.GetDraft() {
			continue
		}
		if !semver.IsValid(release.GetTagName()) {
			continue
		}

		if release.GetPrerelease() {
			prereleases[release.GetTagName()] = struct{}{}
			continue
		}

		// maybe filter alpha/beta?
		candidates = append(candidates, release.GetTagName())
	}

	// If the previous version was a pre-release, consider upgrading to pre-releases:
	if _, ok := prereleases[dependency.Version]; ok {
		log.Debug("including pre-releases")
		for v := range prereleases {
			candidates = append(candidates, v)
		}
	}

	sort.SliceStable(candidates, func(i, j int) bool {
		return semver.Compare(candidates[i], candidates[j]) > 0
	})
	log.WithField("candidates", len(candidates)).Debug("filtered releases")
	return candidates, nil
}
