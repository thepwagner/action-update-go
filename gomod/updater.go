package gomod

import (
	"context"
	"fmt"
	"strings"

	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/modfile"
	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/modload"
	"github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

const (
	GoModFn         = "go.mod"
	VendorModulesFn = "vendor/modules.txt"
)

type Updater struct {
	repo       Repo
	checker    *UpdateChecker
	branchName UpdateBranchNamer

	github   *github.Client
	owner    string
	repoName string

	Tidy   bool
	Author GitIdentity
}

func NewUpdater(repo Repo, ghRepo, token string) (*Updater, error) {
	u := &Updater{
		repo: repo,
		checker: &UpdateChecker{
			MajorVersions: true,
		},
		branchName: DefaultUpdateBranchNamer,
		Tidy:       true,
		Author: GitIdentity{
			Name:  "actions-update-go",
			Email: "noreply@github.com",
		},
	}

	if token != "" {
		ghRepoSplit := strings.Split(ghRepo, "/")
		if len(ghRepoSplit) != 2 {
			return nil, fmt.Errorf("expected repo in OWNER/NAME format")
		}
		u.owner = ghRepoSplit[0]
		u.repoName = ghRepoSplit[1]

		ts := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
		tc := oauth2.NewClient(context.Background(), ts)
		u.github = github.NewClient(tc)
	}

	return u, nil
}

func (u *Updater) UpdateAll(ctx context.Context, branch string) error {
	// Parse go.mod from the branch:
	goMod, err := u.parseGoMod(branch)
	if err != nil {
		return fmt.Errorf("parsing go.mod: %w", err)
	}
	log := logrus.WithField("branch", branch)
	log.WithField("deps", len(goMod.Require)).Info("parsed go.mod, checking for updates")

	// Iterate dependencies:
	modload.Init()
	for _, req := range goMod.Require {
		update, err := u.checker.CheckForModuleUpdates(req)
		if err != nil {
			log.WithError(err).Warn("error checking for updates")
			continue
		}
		if update == nil {
			continue
		}

		if err := u.update(ctx, branch, *update); err != nil {
			return fmt.Errorf("updating %q: %w", update.Path, err)
		}
	}
	return nil
}

func (u *Updater) parseGoMod(branch string) (*modfile.File, error) {
	b, err := u.repo.ReadFile(branch, GoModFn)
	if err != nil {
		return nil, fmt.Errorf("opening go.mod: %w", err)
	}
	parsed, err := modfile.Parse(GoModFn, b, nil)
	if err != nil {
		return nil, fmt.Errorf("parsing go.mod: %w", err)
	}
	return parsed, nil
}

func pathMajorVersion(pkg string, version int64) string {
	return fmt.Sprintf("%s/v%d", pkg[:strings.LastIndex(pkg, "/")], version)
}

func (u *Updater) update(ctx context.Context, baseBranch string, update ModuleUpdate) error {
	targetBranch := u.branchName(baseBranch, update)
	sbx, err := u.repo.NewSandbox(baseBranch, targetBranch)
	if err != nil {
		return fmt.Errorf("preparing update sandbox: %w", err)
	}
	defer sbx.Close()

	if err := UpdateSandbox(ctx, sbx, update, u.Tidy); err != nil {
		return fmt.Errorf("applying update: %w", err)
	}

	// TODO: dependency inject this
	commitMessage := fmt.Sprintf("update %s to %s", update.Path, update.Next)

	if err := sbx.Commit(ctx, commitMessage, u.Author); err != nil {
		return fmt.Errorf("pushing update: %w", err)
	}
	// TODO: create PR
	return nil
}

//	if err := u.createPR(ctx, update); err != nil {
//		return err
//	}
//	return nil
//}
//

//
//func (u *Updater) createPR(ctx context.Context, update ModuleUpdate) error {
//	if u.github == nil {
//		return nil
//	}
//
//	// TODO: dependency inject this
//	title := fmt.Sprintf("Update %s from %s to %s", update.Path, update.Previous, update.Next)
//	var body strings.Builder
//	_, _ = fmt.Fprintln(&body, "you're welcome")
//	_, _ = fmt.Fprintln(&body, "")
//	_, _ = fmt.Fprintln(&body, "TODO: release notes or something?")
//	_, _ = fmt.Fprintln(&body, "")
//	_, _ = fmt.Fprintln(&body, "```json")
//	major := semver.Major(update.Previous) != semver.Major(update.Next)
//	minor := !major && semver.MajorMinor(update.Previous) != semver.MajorMinor(update.Next)
//	details := struct {
//		Major bool `json:"Major"`
//		Minor bool `json:"minor"`
//		Patch bool `json:"patch"`
//	}{
//		Major: major,
//		Minor: minor,
//		Patch: !major && !minor,
//	}
//	encoder := json.NewEncoder(&body)
//	encoder.SetIndent("", "  ")
//	if err := encoder.Encode(&details); err != nil {
//		return err
//	}
//	_, _ = fmt.Fprintln(&body, "")
//	_, _ = fmt.Fprintln(&body, "```")
//
//	res, _, err := u.github.PullRequests.Create(ctx, u.owner, u.repoName, &github.NewPullRequest{
//		Title: &title,
//		Body:  github.String(body.String()),
//		Base:  github.String(update.Base.Name().Short()),
//		Head:  github.String(update.BranchName()),
//	})
//	if err != nil {
//		if strings.Contains(err.Error(), "pull request already exists") {
//			return nil
//		}
//		return fmt.Errorf("creating PR: %w", err)
//	}
//	logrus.WithField("pr_id", res.ID).Info("created pull request")
//	return nil
//}
