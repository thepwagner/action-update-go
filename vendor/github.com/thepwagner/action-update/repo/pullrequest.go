package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/google/go-github/v35/github"
	"github.com/thepwagner/action-update/updater"
)

type GitHubPullRequestContent struct {
	github *github.Client
	key    []byte
}

var _ PullRequestContent = (*GitHubPullRequestContent)(nil)

func NewGitHubPullRequestContent(gh *github.Client, key []byte) *GitHubPullRequestContent {
	return &GitHubPullRequestContent{
		github: gh,
		key:    key,
	}
}

func (d *GitHubPullRequestContent) Generate(ctx context.Context, updates updater.UpdateGroup) (title, body string, err error) {
	if len(updates.Updates) == 1 {
		update := updates.Updates[0]
		title = fmt.Sprintf("Update %s from %s to %s", update.Path, update.Previous, update.Next)
		body, err = d.bodySingle(ctx, updates)
	} else {
		title = "Dependency Updates"
		body, err = d.bodyMulti(ctx, updates)
	}
	return
}

const (
	openToken  = "<!--::action-update-go::"
	closeToken = "-->"
)

func (d *GitHubPullRequestContent) ParseBody(s string) *updater.UpdateGroup {
	signed := ExtractSignedUpdateGroup(s)
	if signed == nil {
		return nil
	}

	updates, _ := updater.VerifySignedUpdateGroup(d.key, *signed)
	return updates
}

func ExtractSignedUpdateGroup(s string) *updater.SignedUpdateGroup {
	lastOpen := strings.LastIndex(s, openToken)
	if lastOpen == -1 {
		return nil
	}
	closeAfterOpen := strings.Index(s[lastOpen:], closeToken)
	raw := s[lastOpen+len(openToken) : lastOpen+closeAfterOpen]

	var signed updater.SignedUpdateGroup
	if err := json.Unmarshal([]byte(raw), &signed); err != nil {
		return nil
	}
	return &signed
}

func (d *GitHubPullRequestContent) bodySingle(ctx context.Context, updates updater.UpdateGroup) (string, error) {
	update := updates.Updates[0]
	var body strings.Builder
	_, _ = fmt.Fprintf(&body, "Here is %s %s, I hope it works.\n", update.Path, update.Next)

	if err := d.writeGitHubChangelog(ctx, &body, update); err != nil {
		return "", err
	}

	if err := d.writeUpdateSignature(&body, updates); err != nil {
		return "", fmt.Errorf("writing update signature: %w", err)
	}
	return body.String(), nil
}

func (d *GitHubPullRequestContent) bodyMulti(ctx context.Context, updates updater.UpdateGroup) (string, error) {
	var body strings.Builder
	body.WriteString("Here are some updates, I hope they work.\n\n")

	for _, update := range updates.Updates {
		_, _ = fmt.Fprintf(&body, "#### %s@%s\n", update.Path, update.Next)
		before := body.Len()
		if err := d.writeGitHubChangelog(ctx, &body, update); err != nil {
			return "", err
		}
		if body.Len() != before {
			body.WriteString("\n")
		}
	}

	if err := d.writeUpdateSignature(&body, updates); err != nil {
		return "", fmt.Errorf("writing update signature: %w", err)
	}
	return body.String(), nil
}

func (d *GitHubPullRequestContent) writeGitHubChangelog(ctx context.Context, out io.Writer, update updater.Update) error {
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

func (d *GitHubPullRequestContent) writeUpdateSignature(out io.Writer, updates updater.UpdateGroup) error {
	dsc, err := updater.NewSignedUpdateGroup(d.key, updates)
	if err != nil {
		return fmt.Errorf("signing updates: %w", err)
	}

	_, _ = fmt.Fprint(out, "\n", openToken, "\n")
	if err := json.NewEncoder(out).Encode(&dsc); err != nil {
		return fmt.Errorf("encoding signature: %w", err)
	}
	_, _ = fmt.Fprint(out, closeToken)
	return nil
}
