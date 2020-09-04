package repo

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/semver"
	"github.com/thepwagner/action-update-go/gomod"
)

type DefaultPullRequestContentGenerator struct {
	
}

func NewDefaultPullRequestContentGenerator() *DefaultPullRequestContentGenerator {
	return &DefaultPullRequestContentGenerator{}
}

func (d *DefaultPullRequestContentGenerator) Generate(update gomod.Update) (title, body string, err error) {
	title = fmt.Sprintf("Update %s from %s to %s", update.Path, update.Previous, update.Next)
	body = stubPRBody(update)
	return
}

func stubPRBody(update gomod.Update) string {
	var body strings.Builder
	_, _ = fmt.Fprintln(&body, "you're welcome")
	_, _ = fmt.Fprintln(&body, "")
	_, _ = fmt.Fprintln(&body, "TODO: release notes or something?")
	_, _ = fmt.Fprintln(&body, "")

	major := semver.Major(update.Previous) != semver.Major(update.Next)
	minor := !major && semver.MajorMinor(update.Previous) != semver.MajorMinor(update.Next)
	details := struct {
		Major bool `json:"major"`
		Minor bool `json:"minor"`
		Patch bool `json:"patch"`
	}{
		Major: major,
		Minor: minor,
		Patch: !major && !minor,
	}
	encoder := json.NewEncoder(&body)
	encoder.SetIndent("", "  ")
	_, _ = fmt.Fprintln(&body, "```json")
	_ = encoder.Encode(&details)
	_, _ = fmt.Fprintln(&body, "")
	_, _ = fmt.Fprintln(&body, "```")

	return body.String()
}
