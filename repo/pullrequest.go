package repo

import (
	"context"
	"crypto/hmac"
	"crypto/sha512"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"

	"github.com/google/go-github/v32/github"
	"github.com/thepwagner/action-update-go/updater"
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

func (d *GitHubPullRequestContent) Generate(ctx context.Context, updates ...updater.Update) (title, body string, err error) {
	if len(updates) == 1 {
		update := updates[0]
		title = fmt.Sprintf("Update %s from %s to %s", update.Path, update.Previous, update.Next)
		body, err = d.bodySingle(ctx, update)
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

func (d *GitHubPullRequestContent) ParseBody(s string) []updater.Update {
	lastOpen := strings.LastIndex(s, openToken)
	if lastOpen == -1 {
		return nil
	}
	closeAfterOpen := strings.Index(s[lastOpen:], closeToken)
	raw := s[lastOpen+len(openToken) : lastOpen+closeAfterOpen]

	var signed SignedUpdateDescriptor
	if err := json.Unmarshal([]byte(raw), &signed); err != nil {
		return nil
	}

	updates, _ := VerifySignedUpdateDescriptor(d.key, signed)
	return updates
}

func (d *GitHubPullRequestContent) bodySingle(ctx context.Context, update updater.Update) (string, error) {
	var body strings.Builder
	_, _ = fmt.Fprintf(&body, "Here is %s %s, I hope it works.\n", update.Path, update.Next)

	if err := d.writeGitHubChangelog(ctx, &body, update); err != nil {
		return "", err
	}
	if err := d.writeUpdateSignature(&body, update); err != nil {
		return "", fmt.Errorf("writing update signature: %w", err)
	}
	return body.String(), nil
}

func (d *GitHubPullRequestContent) bodyMulti(ctx context.Context, updates []updater.Update) (string, error) {
	var body strings.Builder
	body.WriteString("Here are some updates, I hope they work.\n\n")

	for _, update := range updates {
		_, _ = fmt.Fprintf(&body, "#### %s@%s\n", update.Path, update.Next)
		before := body.Len()
		if err := d.writeGitHubChangelog(ctx, &body, update); err != nil {
			return "", err
		}
		if body.Len() != before {
			body.WriteString("\n")
		}
	}

	if err := d.writeUpdateSignature(&body, updates...); err != nil {
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

func (d *GitHubPullRequestContent) writeUpdateSignature(out io.Writer, updates ...updater.Update) error {
	dsc, err := NewSignedUpdateDescriptor(d.key, updates...)
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

type SignedUpdateDescriptor struct {
	Updates   []updater.Update `json:"updates"`
	Signature []byte           `json:"signature"`
}

func NewSignedUpdateDescriptor(key []byte, updates ...updater.Update) (SignedUpdateDescriptor, error) {
	signature, err := updatesHash(key, updates)
	if err != nil {
		return SignedUpdateDescriptor{}, err
	}
	return SignedUpdateDescriptor{
		Updates:   updates,
		Signature: signature,
	}, nil
}

func updatesHash(key []byte, updates []updater.Update) ([]byte, error) {
	sort.Slice(updates, func(i, j int) bool {
		return updates[i].Path < updates[j].Path
	})
	hash := hmac.New(sha512.New, key)
	if err := json.NewEncoder(hash).Encode(updates); err != nil {
		return nil, err
	}
	return hash.Sum(nil), nil
}

func VerifySignedUpdateDescriptor(key []byte, descriptor SignedUpdateDescriptor) ([]updater.Update, error) {
	calculated, err := updatesHash(key, descriptor.Updates)
	if err != nil {
		return nil, fmt.Errorf("calculating signature: %w", err)
	}
	if subtle.ConstantTimeCompare(calculated, descriptor.Signature) != 1 {
		return nil, fmt.Errorf("invalid signature")
	}
	return descriptor.Updates, nil
}
