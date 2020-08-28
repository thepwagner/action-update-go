package api

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/thepwagner/jithub/gh"
)

type Handler struct {
}

func NewHandler() http.Handler {
	h := Handler{}
	return h
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Is this a maven repository request?
	mvnPackage := MavenPackageFromURL(r.URL.Path)
	if mvnPackage == nil {
		logrus.WithField("path", r.URL.Path).Debug("not a maven package")
		http.NotFound(w, r)
		return
	}

	// Is this a repository we care about?
	parts := strings.Split(mvnPackage.GroupID, ".")
	if len(parts) != 3 || parts[0] != "com" || parts[1] != "github" {
		logrus.WithField("path", r.URL.Path).Debug("unknown maven package")
		http.NotFound(w, r)
		return
	}
	owner := parts[2]

	// Authenticated GitHub client:
	// TODO: be an app - blocked by https://github.com/github/ecosystem-apps/issues/727
	ctx := r.Context()
	client := gh.NewStaticTokenClient(ctx, os.Getenv("GITHUB_TOKEN"))

	// Does this package exist?
	files, err := client.PackageFiles(ctx, owner, mvnPackage.ArtifactID, mvnPackage.Version)
	if err != nil {
		logrus.WithError(err).Error("fetching package files")
		var status int
		if strings.Contains(err.Error(), "Could not resolve to a Repository") {
			status = http.StatusNotFound
		} else {
			status = http.StatusInternalServerError
		}
		w.WriteHeader(status)
		return
	}

	if len(files) == 0 {
		logrus.WithField("path", r.URL.Path).Debug("package not found, attempting build...")
		if err := client.TriggerPackageBuild(ctx, owner, mvnPackage.ArtifactID, mvnPackage.Version); err != nil {
			logrus.WithError(err).Error("could not trigger build")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// We've requested the build, try again later:
		w.Header().Set("Retry-After", "600")
		w.WriteHeader(http.StatusGatewayTimeout)
		_, _ = fmt.Fprint(w, "build requested, please wait")
		return
	}

	url, ok := files[mvnPackage.Filename()]
	if !ok {
		logrus.WithField("path", r.URL.Path).Debug("unknown file")
		http.NotFound(w, r)
		return
	}

	logrus.WithField("path", r.URL.Path).Debug("redirected to GPR")
	http.Redirect(w, r, url, http.StatusFound)
	return
}
