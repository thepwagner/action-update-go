package dockerurl

import (
	"context"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/google/go-github/v32/github"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/sirupsen/logrus"
	"github.com/thepwagner/action-update-go/docker"
	"github.com/thepwagner/action-update-go/updater"
)

func (u *Updater) ApplyUpdate(ctx context.Context, update updater.Update) error {
	_, name := parseGitHubRelease(update.Path)
	nameUpper := strings.ToUpper(name)

	// Potential keys the version/release hash might be stored under:
	versionKeys := []string{
		fmt.Sprintf("%s_VERSION", nameUpper),
		fmt.Sprintf("%s_RELEASE", nameUpper),
	}
	hashKeys := []string{
		fmt.Sprintf("%s_SHASUM", nameUpper),
		fmt.Sprintf("%s_SHA256SUM", nameUpper),
		fmt.Sprintf("%s_SHA512SUM", nameUpper),
		fmt.Sprintf("%s_CHECKSUM", nameUpper),
	}

	return docker.WalkDockerfiles(u.root, func(path string, parsed *parser.Result) error {
		patterns := u.collectPatterns(ctx, parsed, versionKeys, hashKeys, update)
		logrus.WithFields(logrus.Fields{
			"path":     path,
			"patterns": len(patterns),
		}).Debug("collected patterns")
		return updateDockerfile(path, patterns)
	})
}

func (u *Updater) collectPatterns(ctx context.Context, parsed *parser.Result, versionKeys, hashKeys []string, update updater.Update) map[string]string {
	i := docker.NewInterpolation(parsed)
	patterns := map[string]string{}
	var versionHit bool
	for _, k := range versionKeys {
		prev, ok := i.Vars[k]
		if !ok {
			continue
		}
		switch prev {
		case update.Previous:
			logrus.WithFields(logrus.Fields{
				"key":    k,
				"prefix": true,
			}).Debug("identified version key")
			patterns[fmt.Sprintf("%s=%s", k, prev)] = fmt.Sprintf("%s=%s", k, update.Next)
			versionHit = true
		case update.Previous[1:]:
			logrus.WithFields(logrus.Fields{
				"key":    k,
				"prefix": false,
			}).Debug("identified version key")
			patterns[fmt.Sprintf("%s=%s", k, update.Previous[1:])] = fmt.Sprintf("%s=%s", k, update.Next[1:])
			versionHit = true
		}
	}
	if !versionHit {
		return patterns
	}

	for _, k := range hashKeys {
		prev, ok := i.Vars[k]
		if !ok {
			continue
		}
		log := logrus.WithField("key", k)
		log.Debug("identified hash key")

		newHash, err := u.updatedHash(ctx, update, prev)
		if err != nil {
			log.WithError(err).Warn("fetching updated hash")
		} else if newHash != "" {
			log.Debug("updated hash key")
			patterns[fmt.Sprintf("%s=%s", k, prev)] = fmt.Sprintf("%s=%s", k, newHash)
		}
	}
	return patterns
}

func (u *Updater) updatedHash(ctx context.Context, update updater.Update, oldHash string) (string, error) {
	// Fetch the previous release:
	owner, repoName := parseGitHubRelease(update.Path)
	prevRelease, _, err := u.ghRepos.GetReleaseByTag(ctx, owner, repoName, update.Previous)
	if err != nil {
		return "", fmt.Errorf("fetching previous release: %w", err)
	}

	// First pass, does the project release a SHASUMS etc file we can grab?
	for _, prevAsset := range prevRelease.Assets {
		log := logrus.WithField("name", prevAsset.GetName())
		if h, err := u.isShasumAsset(ctx, prevAsset, oldHash); err != nil {
			log.WithError(err).Warn("inspecting potential hash asset")
			continue
		} else if !h {
			continue
		}
		log.Debug("identified shasum asset in previous release")

		// The previous release contained a shasum file that contained the previous hash
		// Does the new release have the same file?
		newHash, err := u.updatedHashFromShasumAsset(ctx, prevAsset, update)
		if err != nil {
			log.WithError(err).Warn("fetching updated hash asset")
			continue
		}
		if newHash != "" {
			log.Debug("fetched corresponding shasum asset from new release")
			return newHash, nil
		}
	}

	// There are no shasum files - get downloading
	logrus.Debug("shasum file not found, searching files from previous release")
	for _, prevAsset := range prevRelease.Assets {
		log := logrus.WithField("name", prevAsset.GetName())
		h, err := u.isHashAsset(ctx, prevAsset, oldHash)
		if err != nil {
			log.WithError(err).Warn("checking hash of previous assets")
			continue
		} else if !h {
			continue
		}
		log.Debug("identified hashed asset in previous release")

		// This asset from a previous release matched the previous hash
		// Does the new release have the same file?
		newHash, err := u.updatedHashFromAsset(ctx, prevAsset, update, oldHash)
		if err != nil {
			return "", err
		}
		if newHash != "" {
			log.Debug("fetched corresponding asset from new release")
			return newHash, nil
		}
	}

	return "", nil
}

// isShasumAsset returns true if the release asset is a SHASUMS file containing the previous hash
func (u *Updater) isShasumAsset(ctx context.Context, asset *github.ReleaseAsset, oldHash string) (bool, error) {
	if asset.GetSize() > 1024 {
		return false, nil
	}

	req, err := http.NewRequest("GET", asset.GetBrowserDownloadURL(), nil)
	if err != nil {
		return false, err
	}
	req = req.WithContext(ctx)

	res, err := u.http.Do(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return false, err
	}
	return strings.Contains(string(b), oldHash), nil
}

func (u *Updater) updatedHashFromShasumAsset(ctx context.Context, asset *github.ReleaseAsset, update updater.Update) (string, error) {
	res, err := u.getUpdatedAsset(ctx, asset, update)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	return strings.SplitN(string(b), " ", 2)[0], nil
}

func (u *Updater) getUpdatedAsset(ctx context.Context, asset *github.ReleaseAsset, update updater.Update) (*http.Response, error) {
	newURL := asset.GetBrowserDownloadURL()
	newURL = strings.ReplaceAll(newURL, update.Previous, update.Next)
	newURL = strings.ReplaceAll(newURL, update.Previous[1:], update.Next[1:])
	req, err := http.NewRequest("GET", newURL, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)
	return u.http.Do(req)
}

func (u *Updater) isHashAsset(ctx context.Context, asset *github.ReleaseAsset, oldHash string) (bool, error) {
	h, ok := hasher(oldHash)
	if !ok {
		return false, nil
	}

	req, err := http.NewRequest("GET", asset.GetBrowserDownloadURL(), nil)
	if err != nil {
		return false, err
	}
	req = req.WithContext(ctx)

	res, err := u.http.Do(req)
	if err != nil {
		return false, err
	}
	defer res.Body.Close()
	if _, err := io.Copy(h, res.Body); err != nil {
		return false, err
	}
	sum := h.Sum(nil)
	return fmt.Sprintf("%x", sum) == oldHash, nil
}

func hasher(oldHash string) (hash.Hash, bool) {
	switch len(oldHash) {
	case 64:
		return sha256.New(), true
	case 128:
		return sha512.New(), true
	default:
		return nil, false
	}
}

func (u *Updater) updatedHashFromAsset(ctx context.Context, asset *github.ReleaseAsset, update updater.Update, oldHash string) (string, error) {
	res, err := u.getUpdatedAsset(ctx, asset, update)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()
	h, _ := hasher(oldHash)
	if _, err := io.Copy(h, res.Body); err != nil {
		return "", err
	}
	sum := h.Sum(nil)
	return fmt.Sprintf("%x", sum), nil
}

func updateDockerfile(path string, patterns map[string]string) error {
	// Buffer contents as a string
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return fmt.Errorf("reading file: %w", err)
	}
	s := string(b)

	for old, replace := range patterns {
		s = strings.ReplaceAll(s, old, replace)
	}

	stat, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stating file: %w", err)
	}
	if err := ioutil.WriteFile(path, []byte(s), stat.Mode()); err != nil {
		return fmt.Errorf("writing updated file: %w", err)
	}
	return nil
}
