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
	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/modfile"
	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/modinfo"
	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/semver"
	"github.com/sirupsen/logrus"
)

type UpdateChecker struct {
	MajorVersions bool
	RootDir       string
}

// TODO: support replace directives
// TODO: filtering allowable updates:
// - include/exclude paths to check for updates
// - filter versions in include filter? (e.g. include pkg/foo < v4)

func (c *UpdateChecker) CheckForModuleUpdates(ctx context.Context, req *modfile.Require) (*Update, error) {
	path := req.Mod.Path
	log := logrus.WithField("path", path)

	if modfetch.IsPseudoVersion(req.Mod.Version) {
		log.WithField("version", req.Mod.Version).Debug("skipping psuedoversion module")
		return nil, nil
	}

	if c.MajorVersions {
		latest, err := c.checkForMajorUpdate(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("checking for major update: %w", err)
		}
		if latest != nil {
			return latest, nil
		}
	}

	latest, err := c.checkForUpdate(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("checking for update: %w", err)
	}
	return latest, nil
}

var pathMajorVersionRE = regexp.MustCompile("/v([0-9]+)$")

func pathMajorVersion(path string, version int64) string {
	// FIXME: this won't play nice with `github/v32/github`
	return fmt.Sprintf("%s/v%d", path[:strings.LastIndex(path, "/")], version)
}

func (c *UpdateChecker) checkForMajorUpdate(ctx context.Context, req *modfile.Require) (*Update, error) {
	// Does this look like a versioned path?
	path := req.Mod.Path
	m := pathMajorVersionRE.FindStringSubmatch(path)
	if len(m) == 0 {
		return nil, nil
	}
	currentMajorVersion, _ := strconv.ParseInt(m[1], 10, 32)

	log := logrus.WithField("path", path)
	log.Debug("querying latest major version")

	nfo, err := c.queryModuleVersions(ctx, pathMajorVersion(path, currentMajorVersion+1))
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") {
			// Assume we queried for a major version that doesn't exist
			return nil, nil
		}
		return nil, err
	}

	version := req.Mod.Version
	log = log.WithFields(logrus.Fields{
		"latest_version":  nfo.Version,
		"current_version": version,
	})
	if modfetch.IsPseudoVersion(nfo.Version) {
		log.Debug("skipping major update to pseudoversion")
		return nil, nil
	}

	log.Info("major upgrade available")
	return &Update{
		Path:     path,
		Previous: version,
		Next:     nfo.Version,
	}, nil
}

func (c *UpdateChecker) checkForUpdate(ctx context.Context, req *modfile.Require) (*Update, error) {
	path := req.Mod.Path
	log := logrus.WithField("path", path)
	log.Debug("querying latest version")

	nfo, err := c.queryModuleVersions(ctx, path)
	if err != nil {
		return nil, err
	}

	var latestVersion = nfo.Version
	if versCount := len(nfo.Versions); versCount > 0 {
		latestVersion = nfo.Versions[versCount-1]
	}

	// Does this update progress the semver?
	version := req.Mod.Version
	log = log.WithFields(logrus.Fields{
		"latest_version":  latestVersion,
		"current_version": version,
	})
	if upgrade := semver.Compare(version, latestVersion) < 0; !upgrade {
		log.Debug("no update available")
		return nil, nil
	}
	log.Info("update available")
	return &Update{
		Path:     path,
		Previous: version,
		Next:     latestVersion,
	}, nil
}

func (c *UpdateChecker) queryModuleVersions(ctx context.Context, nextVersionPath string) (*modinfo.ModulePublic, error) {
	var buf bytes.Buffer
	var errBuf bytes.Buffer
	cmd := exec.CommandContext(ctx, "go", "list", "-m", "-mod=mod", "-versions", "-json", nextVersionPath)
	cmd.Stdout = &buf
	cmd.Stderr = &errBuf
	cmd.Dir = c.RootDir
	if err := cmd.Run(); err != nil {
		logrus.WithField("stderr", errBuf.String()).Warn("module versions query error")
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
