package gomod

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"time"

	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/modfile"
	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/modload"
	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/semver"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/sirupsen/logrus"
)

const (
	goModFn         = "go.mod"
	vendorModulesFn = "vendor/modules.txt"
)

type Updater struct {
	repo      *git.Repository
	wt        *git.Worktree
	Tidy      bool
	Author    GitIdentity
	Committer GitIdentity
}

type GitIdentity struct {
	Name  string
	Email string
}

func NewUpdater(repo *git.Repository) (*Updater, error) {
	wt, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("getting work tree: %w", err)
	}

	if status, err := wt.Status(); err != nil {
		return nil, fmt.Errorf("getting worktree status: %w", err)
	} else if !status.IsClean() {
		return nil, fmt.Errorf("tree is not clean, reset") // or implement force...
	}
	return &Updater{
		repo: repo,
		wt:   wt,
		Tidy: true,
		Author: GitIdentity{
			Name:  "actions-update-go",
			Email: "noreply@github.com",
		},
		Committer: GitIdentity{
			Name:  "actions-update-go",
			Email: "noreply@github.com",
		},
	}, nil
}

func (u *Updater) UpdateAll(baseBranch string) error {
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
		latest, err := u.checkForUpdate(req)
		if err != nil {
			log.WithError(err).Warn("error checking for updates")
			continue
		}
		if latest == "" {
			continue
		}

		if err := u.update(baseRef.Hash(), pkg, latest); err != nil {
			return fmt.Errorf("upgrading %q: %w", pkg, err)
		}
	}
	return nil
}

func (u *Updater) parseGoMod() (*modfile.File, error) {
	f, err := u.wt.Filesystem.Open(goModFn)
	if err != nil {
		return nil, fmt.Errorf("opening go.mod: %w", err)
	}
	defer f.Close()

	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("reading go.mod: %w", err)
	}
	parsed, err := modfile.Parse(goModFn, b, nil)
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
	log = log.WithFields(logrus.Fields{
		"latest_version":  latest.Version,
		"current_version": version,
	})

	upgrade := semver.Compare(version, latest.Version) < 0
	if !upgrade {
		log.Debug("no update available")
		return "", nil
	}
	log.Info("upgrade available")
	return latest.Version, nil
}

func (u *Updater) update(base plumbing.Hash, pkg, version string) error {
	if err := u.createUpdateBranch(base, pkg, version); err != nil {
		return err
	}

	if err := u.updateFiles(pkg, version); err != nil {
		return err
	}

	commitMessage := fmt.Sprintf("update %s from to %s", pkg, version)
	if err := u.commit(commitMessage); err != nil {
		return err
	}

	return nil
}

func (u *Updater) createUpdateBranch(base plumbing.Hash, pkg, version string) error {
	log := logrus.WithFields(logrus.Fields{
		"pkg":     pkg,
		"version": version,
	})
	// Switch to the target branch:
	branchName := fmt.Sprintf("action-update-go/%s/%s", pkg, version)
	log.WithField("branch", branchName).Debug("checking out target branch")
	branchRef := plumbing.NewBranchReferenceName(branchName)
	err := u.wt.Checkout(&git.CheckoutOptions{
		Branch: branchRef,
		Hash:   base,
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

func (u *Updater) updateFiles(pkg, version string) error {
	if err := u.updateGoMod(pkg, version); err != nil {
		return err
	}
	if err := u.updateGoSum(); err != nil {
		return err
	}

	if u.hasVendor() {
		if err := u.updateVendor(); err != nil {
			return err
		}
	}
	return nil
}

func (u *Updater) updateGoMod(pkg, version string) error {
	goMod, err := u.parseGoMod()
	if err != nil {
		return err
	}
	if err := goMod.AddRequire(pkg, version); err != nil {
		return fmt.Errorf("adding requirement: %w", err)
	}
	updated, err := goMod.Format()
	if err != nil {
		return fmt.Errorf("formatting go.mod: %w", err)
	}
	out, err := u.wt.Filesystem.OpenFile(goModFn, os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening go.mod: %w", err)
	}
	defer out.Close()
	if _, err := out.Write(updated); err != nil {
		return fmt.Errorf("writing updated go.mod: %w", err)
	}
	return nil
}

func (u *Updater) updateGoSum() error {
	// Shell out to the Go SDK for this, so the user has more control over generation:
	if err := u.worktreeCmd("go", "get", "-d", "-v"); err != nil {
		return fmt.Errorf("updating go.sum: %w", err)
	}

	if u.Tidy {
		if err := u.worktreeCmd("go", "mod", "tidy"); err != nil {
			return fmt.Errorf("tidying go.sum: %w", err)
		}
	}

	return nil
}

func (u *Updater) hasVendor() bool {
	_, err := u.wt.Filesystem.Stat(vendorModulesFn)
	return err == nil
}

func (u *Updater) updateVendor() error {
	if err := u.worktreeCmd("go", "mod", "vendor"); err != nil {
		return fmt.Errorf("go vendoring: %w", err)
	}
	return nil
}

func (u *Updater) worktreeCmd(cmd string, args ...string) error {
	var out io.Writer
	if logrus.IsLevelEnabled(logrus.DebugLevel) {
		log := logrus.StandardLogger().WriterLevel(logrus.DebugLevel)
		defer log.Close()
		out = log
		_, _ = fmt.Fprintln(out, append([]string{cmd}, args...))
	} else {
		out = ioutil.Discard
	}

	c := exec.Command(cmd, args...)
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
