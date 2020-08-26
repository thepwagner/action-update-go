package gomod

import (
	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/modload"
	"github.com/sirupsen/logrus"
)

func CurrentVersion(pkgName string) string {
	affectedModules := modload.ListModules([]string{pkgName}, false, false)
	if len(affectedModules) != 1 {
		logrus.Info("package not found")
		return ""
	}
	return affectedModules[0].Version
}
