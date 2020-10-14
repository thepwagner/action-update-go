package gomodules

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/modfetch"
	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/modinfo"
	"github.com/sirupsen/logrus"
	"github.com/thepwagner/action-update/updater"
	"golang.org/x/mod/semver"
)

func (u *Updater) Check(ctx context.Context, dep updater.Dependency, filter func(string) bool) (*updater.Update, error) {
	log := logrus.WithField("path", dep.Path)

	if modfetch.IsPseudoVersion(dep.Version) {
		log.WithField("version", dep.Version).Debug("skipping pseudoversion module")
		return nil, nil
	}

	if u.MajorVersions {
		latest, err := u.checkForMajorUpdate(ctx, dep, filter)
		if err != nil {
			return nil, fmt.Errorf("checking for major update: %w", err)
		}
		if latest != nil {
			return latest, nil
		}
	}

	latest, err := u.checkForUpdate(ctx, dep, filter)
	if err != nil {
		return nil, fmt.Errorf("checking for update: %w", err)
	}
	return latest, nil
}

func (u *Updater) checkForMajorUpdate(ctx context.Context, dep updater.Dependency, filter func(string) bool) (*updater.Update, error) {
	// Does this look like a versioned path?
	nextMajorPath := pathNextMajorVersion(dep.Path)
	if nextMajorPath == "" {
		return nil, nil
	}

	log := logrus.WithField("path", dep.Path)
	log.Debug("querying latest major version")

	latest, err := u.queryModuleVersions(ctx, nextMajorPath, filter)
	if err != nil {
		if strings.Contains(err.Error(), "exit status 1") {
			// Assume we queried for a major version that doesn't exist
			return nil, nil
		}
		return nil, err
	} else if latest == nil {
		return nil, nil
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

var pathMajorVersionRE = regexp.MustCompile(`([\\./])v(\d+)$`)

func pathNextMajorVersion(path string) string {
	m := pathMajorVersionRE.FindStringSubmatch(path)
	if len(m) == 0 {
		return ""
	}
	currentMajorVersion, _ := strconv.ParseInt(m[2], 10, 32)
	sep := m[1]
	return fmt.Sprintf("%s%sv%d", path[:strings.LastIndex(path, sep)], sep, currentMajorVersion+1)
}

func pathMajorVersion(basePath, major string) string {
	m := pathMajorVersionRE.FindStringSubmatch(basePath)
	if len(m) == 0 {
		return ""
	}
	sep := m[1]
	return fmt.Sprintf("%s%s%s", basePath[:strings.LastIndex(basePath, sep)], sep, major)
}

func (u *Updater) checkForUpdate(ctx context.Context, dep updater.Dependency, filter func(string) bool) (*updater.Update, error) {
	log := logrus.WithField("path", dep.Path)
	log.Debug("querying latest version")

	nfo, err := u.queryModuleVersions(ctx, dep.Path, filter)
	if err != nil {
		return nil, err
	} else if nfo == nil {
		return nil, nil
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

func (u *Updater) queryModuleVersions(ctx context.Context, path string, filter func(string) bool) (*modinfo.ModulePublic, error) {
	if closer, err := u.ensureGomodInRoot(); err != nil {
		return nil, err
	} else if closer != nil {
		defer func() {
			if err := closer(); err != nil {
				logrus.WithError(err).Warn("cleaning up temp go file")
			}
		}()
	}

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

	if filter != nil {
		if !filter(nfo.Version) {
			nfo.Version = ""
		}

		filtered := make([]string, 0, len(nfo.Versions))
		for _, v := range nfo.Versions {
			if filter(v) {
				filtered = append(filtered, v)
			}
		}
		nfo.Versions = filtered
	}
	if nfo.Version == "" && len(nfo.Versions) == 0 {
		logrus.WithField("path", nfo.Path).Info("all versions ignored by filter")
		return nil, nil
	}

	return &nfo, nil
}

var dummyModFile = []byte(`module dummy`)

func (u *Updater) ensureGomodInRoot() (func() error, error) {
	// Check for go.mod file
	gomodPath := filepath.Join(u.root, GoModFn)
	_, err := os.Stat(gomodPath)
	if err == nil {
		return nil, nil
	} else if !os.IsNotExist(err) {
		return nil, fmt.Errorf("stat go.mod: %w", err)
	}

	// Not found, write a dummy file
	if err := ioutil.WriteFile(gomodPath, dummyModFile, 0600); err != nil {
		return nil, fmt.Errorf("writing dummy go file")
	}
	return func() error {
		return os.Remove(gomodPath)
	}, nil
}
