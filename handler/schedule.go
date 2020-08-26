package handler

import (
	"context"
	"fmt"

	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/modload"
	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/semver"
	"github.com/sirupsen/logrus"
	"github.com/thepwagner/action-update-go/gomod"
)

func Schedule(context.Context, interface{}) error {
	goMod, err := gomod.Parse()
	if err != nil {
		return err
	}

	modload.InitMod()
	for _, req := range goMod.Require {
		pkg := req.Mod.Path
		version := req.Mod.Version
		log := logrus.WithFields(logrus.Fields{
			"pkg":             pkg,
			"current_version": version,
		})
		log.Debug("checking for updates")

		latest, err := modload.Query(pkg, "latest", nil)
		if err != nil {
			log.WithError(err).Warn("querying for latest version")
			continue
		}
		log = log.WithFields(logrus.Fields{
			"latest_version": latest.Version,
		})

		upgrade := semver.Compare(version, latest.Version) < 0
		if !upgrade {
			log.Debug("no update available")
			continue
		}
		log.Info("upgrade available")

		if err := gomod.Update(pkg, latest.Version); err != nil {
			return fmt.Errorf("upgrading %q: %w", pkg, err)
		}
	}

	return nil
}
