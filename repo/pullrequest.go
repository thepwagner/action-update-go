package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/semver"
	"github.com/google/go-github/v32/github"
	"github.com/thepwagner/action-update-go/gomod"
)

type GitHubPullRequestContent struct {
	github *github.Client
}

var _ PullRequestContent = (*GitHubPullRequestContent)(nil)

func NewGitHubPullRequestContent(gh *github.Client) *GitHubPullRequestContent {
	return &GitHubPullRequestContent{github: gh}
}

func (d *GitHubPullRequestContent) Generate(ctx context.Context, update gomod.Update) (title, body string, err error) {
	title = fmt.Sprintf("Update %s from %s to %s", update.Path, update.Previous, update.Next)
	body, err = d.prBody(ctx, update)
	return
}

func (d *GitHubPullRequestContent) prBody(ctx context.Context, update gomod.Update) (string, error) {
	var body strings.Builder
	_, _ = fmt.Fprintf(&body, "Here is %s %s, I hope it works.\n", update.Path, update.Next)

	if err := d.writeGitHubChangelog(ctx, &body, update); err != nil {
		return "", err
	}
	writePatchBlob(&body, update)
	return body.String(), nil
}

func (d *GitHubPullRequestContent) writeGitHubChangelog(ctx context.Context, out io.Writer, update gomod.Update) error {
	if !strings.HasPrefix(update.Path, "github.com/") {
		return nil
	}

	pathSplit := strings.SplitN(update.Path, "/", 4)
	owner := pathSplit[1]
	repo := pathSplit[2]

	contents, _, _, err := d.github.Repositories.GetContents(ctx, owner, repo, "CHANGELOG.md",
		&github.RepositoryContentGetOptions{Ref: update.Next})
	if err != nil {
		if errRes, ok := err.(*github.ErrorResponse); ok && errRes.Response.StatusCode == http.StatusNotFound {
			return nil
		}
		return err
	}
	_, _ = fmt.Fprintf(out, "\n[changelog](%s)\n", contents.GetHTMLURL())
	return nil
}

func writePatchBlob(out io.Writer, update gomod.Update) {
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
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	_, _ = fmt.Fprintln(out, "\n```json")
	_ = encoder.Encode(&details)
	_, _ = fmt.Fprint(out, "```\n")
}
