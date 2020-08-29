package gomod

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/modfetch"
	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/modfile"
	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/modload"
	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/semver"
	"github.com/sirupsen/logrus"
)

type UpdateChecker struct {
	MajorVersions bool
}

func (c *UpdateChecker) CheckForModuleUpdates(req *modfile.Require) (*ModuleUpdate, error) {
	path := req.Mod.Path
	log := logrus.WithField("path", path)

	if modfetch.IsPseudoVersion(req.Mod.Version) {
		log.WithField("version", req.Mod.Version).Debug("skipping psuedoversion module")
		return nil, nil
	}

	if c.MajorVersions {
		latest, err := checkForMajorUpdate(req)
		if err != nil {
			return nil, fmt.Errorf("checking for major update: %w", err)
		}
		if latest != nil {
			return latest, nil
		}
	}

	latest, err := checkForUpdate(req)
	if err != nil {
		return nil, fmt.Errorf("checking for update: %w", err)
	}
	return latest, nil
}

var pathMajorVersionRE = regexp.MustCompile("/v([0-9]+)$")

func checkForMajorUpdate(req *modfile.Require) (*ModuleUpdate, error) {
	// Does this look like a versioned path?
	path := req.Mod.Path
	m := pathMajorVersionRE.FindStringSubmatch(path)
	if len(m) == 0 {
		return nil, nil
	}
	currentMajorVersion, _ := strconv.ParseInt(m[1], 10, 32)

	version := req.Mod.Version
	log := logrus.WithField("path", path)
	log.Debug("querying latest major version")

	latest, err := modload.Query(pathMajorVersion(path, currentMajorVersion+1), "latest", nil)
	if err != nil {
		return nil, fmt.Errorf("querying for latest version: %w", err)
	}
	log = log.WithFields(logrus.Fields{
		"latest_version":  latest.Version,
		"current_version": version,
	})
	if modfetch.IsPseudoVersion(latest.Version) {
		log.Debug("skipping major update to pseudoversion")
		return nil, nil
	}

	log.Info("major upgrade available")
	return &ModuleUpdate{
		Path:     path,
		Previous: version,
		Next:     latest.Version,
	}, nil
}

func checkForUpdate(req *modfile.Require) (*ModuleUpdate, error) {
	path := req.Mod.Path
	log := logrus.WithField("path", path)
	log.Debug("querying latest version")

	latest, err := modload.Query(path, "latest", nil)
	if err != nil {
		return nil, fmt.Errorf("querying for latest version: %w", err)
	}

	// Does this update progress the semver?
	version := req.Mod.Version
	log = log.WithFields(logrus.Fields{
		"latest_version":  latest.Version,
		"current_version": version,
	})
	if upgrade := semver.Compare(version, latest.Version) < 0; !upgrade {
		log.Debug("no update available")
		return nil, nil
	}
	log.Info("update available")
	return &ModuleUpdate{
		Path:     path,
		Previous: version,
		Next:     latest.Version,
	}, nil
}
