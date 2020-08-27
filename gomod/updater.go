package gomod

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/modfetch"
	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/modfile"
	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/modload"
	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/semver"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

const (
	GoModFn         = "go.mod"
	VendorModulesFn = "vendor/modules.txt"
)

type Updater struct {
	repo     *git.Repository
	wt       *git.Worktree
	github   *github.Client
	owner    string
	repoName string

	Tidy         bool
	Author       GitIdentity
	Committer    GitIdentity
	MajorVersion bool
}

type GitIdentity struct {
	Name  string
	Email string
}

func NewUpdater(repo *git.Repository, ghRepo, token string) (*Updater, error) {
	wt, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("getting work tree: %w", err)
	}
	if status, err := wt.Status(); err != nil {
		return nil, fmt.Errorf("getting worktree status: %w", err)
	} else if !status.IsClean() {
		return nil, fmt.Errorf("tree is not clean, reset") // or implement force...
	}

	u := &Updater{
		repo: repo,
		wt:   wt,

		Tidy:         true,
		MajorVersion: true,
		Author: GitIdentity{
			Name:  "actions-update-go",
			Email: "noreply@github.com",
		},
		Committer: GitIdentity{
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

func (u *Updater) UpdateAll(ctx context.Context, baseBranch string) error {
	log := logrus.WithField("branch", baseBranch)

	// Switch to the target branch:
	log.Debug("checking out base branch")
	baseBranchRef := plumbing.NewBranchReferenceName(baseBranch)
	err := u.wt.Checkout(&git.CheckoutOptions{
		Branch: baseBranchRef,
		Force:  true,
	})
	if err != nil {
		return fmt.Errorf("switching to target branch: %w", err)
	}
	baseRef, err := u.repo.Reference(baseBranchRef, true)
	if err != nil {
		return fmt.Errorf("getting target branch hash: %w", err)
	}

	// Parse go.mod to list direct dependencies:
	goMod, err := u.parseGoMod()
	if err != nil {
		return fmt.Errorf("parsing go.mod: %w", err)
	}
	log.WithField("deps", len(goMod.Require)).Info("checking for updates...")

	modload.Init()
	for _, req := range goMod.Require {
		pkg := req.Mod.Path
		log := logrus.WithField("pkg", pkg)

		if modfetch.IsPseudoVersion(req.Mod.Version) {
			log.WithField("version", req.Mod.Version).Debug("skipping psuedoversion module")
			continue
		}

		update := moduleUpdate{
			baseBranch: baseBranch,
			baseHash:   baseRef.Hash(),
			pkg:        pkg,
			previous:   req.Mod.Version,
		}

		if u.MajorVersion {
			latest, err := u.checkForMajorUpdate(req)
			if err != nil {
				log.WithError(err).Warn("error checking for major update")
				continue
			}
			if latest != "" {
				update.next = latest
				if err := u.update(ctx, update); err != nil {
					return fmt.Errorf("upgrading %q: %w", pkg, err)
				}
				continue
			}
		}

		latest, err := u.checkForUpdate(req)
		if err != nil {
			log.WithError(err).Warn("error checking for update")
			continue
		}
		if latest == "" {
			continue
		}
		update.next = latest
		if err := u.update(ctx, update); err != nil {
			return fmt.Errorf("upgrading %q: %w", pkg, err)
		}
	}
	return nil
}

func (u *Updater) parseGoMod() (*modfile.File, error) {
	f, err := u.wt.Filesystem.Open(GoModFn)
	if err != nil {
		return nil, fmt.Errorf("opening go.mod: %w", err)
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("reading go.mod: %w", err)
	}
	parsed, err := modfile.Parse(GoModFn, b, nil)
	if err != nil {
		return nil, fmt.Errorf("parsing go.mod: %w", err)
	}
	return parsed, nil
}

func (u *Updater) checkForUpdate(req *modfile.Require) (latestVersion string, err error) {
	pkg := req.Mod.Path
	version := req.Mod.Version
	log := logrus.WithField("pkg", pkg)
	log.Debug("querying latest version")

	latest, err := modload.Query(pkg, "latest", nil)
	if err != nil {
		return "", fmt.Errorf("querying for latest version: %w", err)
	}

	// Is this an upgrade?
	log = log.WithFields(logrus.Fields{
		"latest_version":  latest.Version,
		"current_version": version,
	})
	if upgrade := semver.Compare(version, latest.Version) < 0; upgrade {
		log.Info("upgrade available")
		return latest.Version, nil
	}
	log.Debug("no update available")
	return "", nil
}

var pkgMajorVersionRE = regexp.MustCompile("/v([0-9]+)$")

func (u *Updater) checkForMajorUpdate(req *modfile.Require) (latestVersion string, err error) {
	// Does this look like a package
	pkg := req.Mod.Path
	m := pkgMajorVersionRE.FindStringSubmatch(pkg)
	if len(m) == 0 {
		return "", nil
	}
	currentMajorVersion, _ := strconv.ParseInt(m[1], 10, 32)

	version := req.Mod.Version
	log := logrus.WithField("pkg", pkg)
	log.Debug("querying latest major version")

	latest, err := modload.Query(pkgMajorVersion(pkg, currentMajorVersion+1), "latest", nil)
	if err != nil {
		return "", fmt.Errorf("querying for latest version: %w", err)
	}
	if modfetch.IsPseudoVersion(latest.Version) {
		return "", nil
	}

	log.WithFields(logrus.Fields{
		"latest_version":  latest.Version,
		"current_version": version,
	}).Info("major upgrade available")
	return latest.Version, nil
}

func pkgMajorVersion(pkg string, version int64) string {
	return fmt.Sprintf("%s/v%d", pkg[:strings.LastIndex(pkg, "/")], version)
}

func (u *Updater) update(ctx context.Context, update moduleUpdate) error {
	if err := u.createUpdateBranch(update); err != nil {
		return err
	}

	if err := u.updateFiles(ctx, update); err != nil {
		return err
	}

	// TODO: dependency inject this
	commitMessage := fmt.Sprintf("update %s to %s", update.pkg, update.next)
	if err := u.commit(commitMessage); err != nil {
		return err
	}

	if err := u.push(ctx); err != nil {
		return err
	}

	if err := u.createPR(ctx, update); err != nil {
		return err
	}
	return nil
}

func (u *Updater) createUpdateBranch(update moduleUpdate) error {
	log := logrus.WithFields(logrus.Fields{
		"pkg":     update.pkg,
		"version": update.next,
	})
	branchName := update.branchName()
	log.WithField("branch", branchName).Debug("checking out target branch")
	branchRef := plumbing.NewBranchReferenceName(branchName)
	err := u.wt.Checkout(&git.CheckoutOptions{
		Branch: branchRef,
		Hash:   update.baseHash,
		Create: true,
		Force:  true,
	})
	if err != nil {
		return fmt.Errorf("switching to target branch: %w", err)
	}
	err = u.repo.CreateBranch(&config.Branch{
		Name:   branchName,
		Merge:  branchRef,
		Remote: "origin",
	})
	if err != nil {
		return fmt.Errorf("creating target branch: %w", err)
	}
	return nil
}

func (u *Updater) updateFiles(ctx context.Context, update moduleUpdate) error {
	if err := u.updateGoMod(update); err != nil {
		return err
	}

	if update.major() {
		if err := u.updateSourceCode(update); err != nil {
			return err
		}
	}

	if err := u.updateGoSum(ctx); err != nil {
		return err
	}

	if u.hasVendor() {
		if err := u.updateVendor(ctx); err != nil {
			return err
		}
	}
	return nil
}

func (u *Updater) updateGoMod(update moduleUpdate) error {
	goMod, err := u.parseGoMod()
	if err != nil {
		return err
	}

	if update.major() {
		// Replace foo.bar/v2 with foo.bar/v3:
		if err := goMod.DropRequire(update.pkg); err != nil {
			return fmt.Errorf("dropping requirement: %w", err)
		}
		pkgNext := path.Join(path.Dir(update.pkg), semver.Major(update.next))
		if err := goMod.AddRequire(pkgNext, update.next); err != nil {
			return fmt.Errorf("dropping requirement: %w", err)
		}
	} else {
		// Replace the version:
		if err := goMod.AddRequire(update.pkg, update.next); err != nil {
			return fmt.Errorf("adding requirement: %w", err)
		}
	}

	updated, err := goMod.Format()
	if err != nil {
		return fmt.Errorf("formatting go.mod: %w", err)
	}
	if logrus.IsLevelEnabled(logrus.DebugLevel) {
		out := logrus.StandardLogger().WriterLevel(logrus.DebugLevel)
		defer out.Close()
		_, _ = fmt.Fprintln(out, "-- go.mod --")
		_, _ = out.Write(updated)
		fmt.Fprintln(out, "-- /go.mod --")
	}

	out, err := u.wt.Filesystem.OpenFile(GoModFn, os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening go.mod: %w", err)
	}
	defer out.Close()
	if _, err := out.Write(updated); err != nil {
		return fmt.Errorf("writing updated go.mod: %w", err)
	}
	return nil
}

func (u *Updater) updateSourceCode(update moduleUpdate) error {
	// replace foo.bar/v1 with foo.bar/v2 in imports:
	pattern, err := regexp.Compile(strings.ReplaceAll(update.pkg, ".", "\\."))
	if err != nil {
		return err
	}

	pkgNext := path.Join(path.Dir(update.pkg), semver.Major(update.next))
	return filepath.Walk(u.wt.Filesystem.Root(), func(path string, info os.FileInfo, err error) error {
		if err != nil {
			logrus.WithError(err).WithField("path", path).Warn("error accessing path")
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) != ".go" {
			return nil
		}
		if err := u.updateFile(path, info, pattern, pkgNext); err != nil {
			return err
		}

		return nil
	})
}

func (u *Updater) updateFile(path string, info os.FileInfo, pattern *regexp.Regexp, replace string) error {
	f, err := os.OpenFile(path, os.O_RDWR, info.Mode())
	if err != nil {
		return err
	}
	defer f.Close()

	var changed bool
	var importing bool
	in := bufio.NewScanner(f)
	var buf bytes.Buffer
	for in.Scan() {
		line := in.Text()
		if line == "import (" {
			importing = true
		} else if line == ")" && importing {
			importing = false
		}

		if importing || strings.HasPrefix(line, "import") {
			replaced := pattern.ReplaceAllString(line, replace)
			changed = changed || replaced != line
			line = replaced
		}

		_, _ = fmt.Fprintln(&buf, line)
	}
	if err := in.Err(); err != nil {
		return err
	}

	if !changed {
		return nil
	}

	if _, err := f.Seek(0, 0); err != nil {
		return err
	}
	if _, err := io.Copy(f, &buf); err != nil {
		return err
	}
	return nil
}

func (u *Updater) updateGoSum(ctx context.Context) error {
	// Shell out to the Go SDK for this, so the user has more control over generation:
	if err := u.worktreeCmd(ctx, "go", "get", "-d", "-v"); err != nil {
		return fmt.Errorf("updating go.sum: %w", err)
	}

	if u.Tidy {
		if err := u.worktreeCmd(ctx, "go", "mod", "tidy"); err != nil {
			return fmt.Errorf("tidying go.sum: %w", err)
		}
	}

	return nil
}

func (u *Updater) hasVendor() bool {
	_, err := u.wt.Filesystem.Stat(VendorModulesFn)
	return err == nil
}

func (u *Updater) updateVendor(ctx context.Context) error {
	if err := u.worktreeCmd(ctx, "go", "mod", "vendor"); err != nil {
		return fmt.Errorf("go vendoring: %w", err)
	}
	return nil
}

func (u *Updater) worktreeCmd(ctx context.Context, cmd string, args ...string) error {
	var out io.Writer
	if logrus.IsLevelEnabled(logrus.DebugLevel) {
		log := logrus.StandardLogger().WriterLevel(logrus.DebugLevel)
		defer log.Close()
		out = log
		_, _ = fmt.Fprintln(out, append([]string{cmd}, args...))
	} else {
		out = ioutil.Discard
	}

	c := exec.CommandContext(ctx, cmd, args...)
	c.Stdout = out
	c.Stderr = out
	c.Dir = u.wt.Filesystem.Root()
	if err := c.Run(); err != nil {
		return err
	}
	return nil
}

func (u *Updater) commit(message string) error {
	when := time.Now()

	// wt.AddGlob() is attractive, but does not respect .gitignore
	// .Status() respects .gitignore so add file by file:
	status, err := u.wt.Status()
	if err != nil {
		return fmt.Errorf("checking status for add: %w", err)
	}
	for fn := range status {
		if _, err := u.wt.Add(fn); err != nil {
			return fmt.Errorf("adding file %q: %w", fn, err)
		}
	}

	commit, err := u.wt.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  u.Author.Name,
			Email: u.Author.Email,
			When:  when,
		},
		Committer: &object.Signature{
			Name:  u.Committer.Name,
			Email: u.Committer.Email,
			When:  when,
		},
	})
	if err != nil {
		return fmt.Errorf("committing branch: %w", err)
	}
	logrus.WithField("commit", commit.String()).Info("committed upgrade")
	return nil
}

func (u *Updater) push(ctx context.Context) error {
	// go-git supports Push, but not the [http "https://github.com/"] .gitconfig directive that actions/checkout uses for auth
	// we could extract from u.repo.Config().Raw, but who are we trying to impress?
	if err := u.worktreeCmd(ctx, "git", "push", "-f"); err != nil {
		return fmt.Errorf("pushing: %w", err)
	}
	return nil
}

type moduleUpdate struct {
	baseBranch string
	baseHash   plumbing.Hash

	pkg      string
	previous string
	next     string
}

func (u moduleUpdate) major() bool {
	return semver.Major(u.previous) != semver.Major(u.next)
}

func (u moduleUpdate) branchName() string {
	var branchPkg string
	if u.major() {
		branchPkg = path.Dir(u.pkg)
	} else {
		branchPkg = u.pkg
	}
	return fmt.Sprintf("action-update-go/%s/%s", branchPkg, u.next)
}

func (u *Updater) createPR(ctx context.Context, update moduleUpdate) error {
	if u.github == nil {
		return nil
	}

	// TODO: dependency inject this
	title := fmt.Sprintf("Update %s from %s to %s", update.pkg, update.previous, update.next)
	body := "you're welcome\nTODO: release notes or something?"

	res, _, err := u.github.PullRequests.Create(ctx, u.owner, u.repoName, &github.NewPullRequest{
		Title: &title,
		Body:  &body,
		Base:  &update.baseBranch,
		Head:  github.String(update.branchName()),
	})
	if err != nil {
		return fmt.Errorf("creating PR: %w", err)
	}
	logrus.WithField("pr_id", res.ID).Info("created pull request")
	return nil
}
