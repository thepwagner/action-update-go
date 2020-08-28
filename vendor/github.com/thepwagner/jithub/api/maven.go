package api

import (
	"fmt"
	"regexp"
	"strings"
)

const (
	mvnGroupRe    = `(?P<groupID>.+?)`
	mvnArtifactRe = `(?P<artifactID>[^/]+)`
	mvnVersionRe  = `(?P<version>[^/]+)`
	mvnTypeRe     = `(?P<type>pom|module|jar|ear|war|sha512)`
)

var mvnFilenameRe = fmt.Sprintf(`(?P<filename>[^/]+\.%s)`, mvnTypeRe)

var mvnPackageRe = regexp.MustCompile(
	fmt.Sprintf(`^%s/%s/%s/%s$`, mvnGroupRe, mvnArtifactRe, mvnVersionRe, mvnFilenameRe),
)

type MavenPackage struct {
	GroupID    string
	ArtifactID string
	Version    string
	Classifier string
	Type       string
}

func (m MavenPackage) Filename() string {
	return fmt.Sprintf("%s-%s.%s", m.ArtifactID, m.Version, m.Type)
}

func MavenPackageFromURL(s string) *MavenPackage {
	var p MavenPackage
	var fn string
	matches := mvnPackageRe.FindStringSubmatch(s)
	if len(matches) == 0 {
		return nil
	}

	for i, s := range mvnPackageRe.SubexpNames() {
		switch s {
		case "groupID":
			p.GroupID = strings.ReplaceAll(matches[i][1:], "/", ".")
		case "artifactID":
			p.ArtifactID = matches[i]
		case "version":
			p.Version = matches[i]
		case "filename":
			fn = matches[i]
		case "type":
			p.Type = matches[i]
		}
	}

	// Sanity check
	if !strings.HasPrefix(fn, fmt.Sprintf("%s-%s", p.ArtifactID, p.Version)) {
		return nil
	}
	// TODO: extract classifier from fn
	return &p
}
