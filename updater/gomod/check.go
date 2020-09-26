package gomod

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/modfetch"
	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/modinfo"
	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/semver"
	"github.com/sirupsen/logrus"
	"github.com/thepwagner/action-update-go/updater"
)

func (u *Updater) Check(ctx context.Context, dep updater.Dependency) (*updater.Update, error) {
	log := logrus.WithField("path", dep.Path)

	if modfetch.IsPseudoVersion(dep.Version) {
		log.WithField("version", dep.Version).Debug("skipping pseudoversion module")
		return nil, nil
	}

	if u.MajorVersions {
		latest, err := u.checkForMajorUpdate(ctx, dep)
		if err != nil {
			return nil, fmt.Errorf("checking for major update: %w", err)
		}
		if latest != nil {
			return latest, nil
		}
	}

	latest, err := u.checkForUpdate(ctx, dep)
	if err != nil {
		return nil, fmt.Errorf("checking for update: %w", err)
	}
	return latest, nil
}

var pathMajorVersionRE = regexp.MustCompile(`/v(\d+)$`)

func pathMajorVersion(path string, version int64) string {
	return fmt.Sprintf("%s/v%d", path[:strings.LastIndex(path, "/")], version)
}

func (u *Updater) checkForMajorUpdate(ctx context.Context, dep updater.Dependency) (*updater.Update, error) {
	// Does this look like a versioned path?
	m := pathMajorVersionRE.FindStringSubmatch(dep.Path)
	if len(m) == 0 {
		return nil, nil
	}
	currentMajorVersion, _ := strconv.ParseInt(m[1], 10, 32)

	log := logrus.WithField("path", dep.Path)
	log.Debug("querying latest major version")

	latest, err := u.queryModuleVersions(ctx, pathMajorVersion(dep.Path, currentMajorVersion+1))
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") {
			// Assume we queried for a major version that doesn't exist
			return nil, nil
		}
		return nil, err
	}

	log = log.WithFields(logrus.Fields{
		"latest_version":  latest.Version,
		"current_version": dep.Version,
	})
	if modfetch.IsPseudoVersion(latest.Version) {
		log.Debug("skipping major update to pseudoversion")
		return nil, nil
	}

	log.Info("major upgrade available")
	return &updater.Update{
		Path:     dep.Path,
		Previous: dep.Version,
		Next:     latest.Version,
	}, nil
}

func (u *Updater) checkForUpdate(ctx context.Context, dep updater.Dependency) (*updater.Update, error) {
	log := logrus.WithField("path", dep.Path)
	log.Debug("querying latest version")

	nfo, err := u.queryModuleVersions(ctx, dep.Path)
	if err != nil {
		return nil, err
	}

	var latestVersion = nfo.Version
	if versCount := len(nfo.Versions); versCount > 0 {
		latestVersion = nfo.Versions[versCount-1]
	}

	// Does this update progress the semver?
	log = log.WithFields(logrus.Fields{
		"latest_version":  latestVersion,
		"current_version": dep.Version,
	})
	if upgrade := semver.Compare(dep.Version, latestVersion) < 0; !upgrade {
		log.Debug("no update available")
		return nil, nil
	}

	log.Info("update available")
	return &updater.Update{
		Path:     dep.Path,
		Previous: dep.Version,
		Next:     latestVersion,
	}, nil
}

func (u *Updater) queryModuleVersions(ctx context.Context, path string) (*modinfo.ModulePublic, error) {
	// Shell out to `go list` for the query, as this supports the same authentication the user's using for `go get`
	var buf bytes.Buffer
	var errBuf bytes.Buffer
	cmd := exec.CommandContext(ctx, "go", "list", "-m", "-mod=mod", "-versions", "-json", path)
	cmd.Stdout = &buf
	cmd.Stderr = &errBuf
	cmd.Dir = u.root
	if err := cmd.Run(); err != nil {
		errString := errBuf.String()
		if !strings.Contains(errString, "no matching versions for query") {
			logrus.WithField("stderr", errString).Warn("module versions query error")
		}
		return nil, fmt.Errorf("querying versions: %w", err)
	}
	var nfo modinfo.ModulePublic
	if err := json.NewDecoder(&buf).Decode(&nfo); err != nil {
		return nil, fmt.Errorf("decoding version query: %w", err)
	}
	if nfo.Version == "" {
		return nil, fmt.Errorf("invalid version response")
	}
	return &nfo, nil
}
