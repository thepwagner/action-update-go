package gitrepo

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/sirupsen/logrus"
)

// SharedSandbox is an update sandbox using a locked git working tree
type SharedSandbox struct {
	lock *sync.Mutex
	wt   *git.Worktree
}

// NewSharedSandbox creates SharedSandbox
func NewSharedSandbox(mu *sync.Mutex, repo *git.Repository, baseBranch string, targetBranch string) (*SharedSandbox, error) {
	mu.Lock()
	// passed to the returned SharedSandbox, which will Unlock() in Close(). earlier exits must Unlock()
	keepMu := false
	defer func() {
		if !keepMu {
			mu.Unlock()
		}
	}()

	logrus.WithFields(logrus.Fields{
		"base":   baseBranch,
		"target": targetBranch,
	}).Debug("creating sandbox")
	baseBranchRef, err := ensureBranchExists(repo, baseBranch)
	if err != nil {
		return nil, err
	}

	wt, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("getting work tree: %w", err)
	}

	targetBranchRefName := plumbing.NewBranchReferenceName(targetBranch)
	err = wt.Checkout(&git.CheckoutOptions{
		Branch: targetBranchRefName,
		Hash:   baseBranchRef.Hash(),
		Create: true,
		Force:  true,
	})
	if err != nil {
		return nil, fmt.Errorf("checking out target branch: %w", err)
	}
	err = repo.CreateBranch(&config.Branch{
		Name:   targetBranch,
		Merge:  targetBranchRefName,
		Remote: RemoteName,
	})
	if err != nil {
		return nil, fmt.Errorf("creating target branch: %w", err)
	}

	keepMu = true
	return &SharedSandbox{
		lock: mu,
		wt:   wt,
	}, nil
}

// Close frees this sandbox.
func (s *SharedSandbox) Close() error {
	s.lock.Unlock()
	return nil
}

func (s *SharedSandbox) WriteFile(path string, data []byte) error {
	out, err := s.wt.Filesystem.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening file %q: %w", path, err)
	}
	defer out.Close()
	if _, err := out.Write(data); err != nil {
		return fmt.Errorf("writing file %q: %w", path, err)
	}
	return nil
}

func (s *SharedSandbox) Cmd(ctx context.Context, cmd string, args ...string) error {
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
	c.Dir = s.wt.Filesystem.Root()
	if err := c.Run(); err != nil {
		return err
	}
	return nil
}

func (s *SharedSandbox) Walk(walkFn filepath.WalkFunc) error {
	root := s.wt.Filesystem.Root()
	return filepath.Walk(root, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(root, p)
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if strings.HasPrefix(rel, ".git/") {
			return nil
		}

		return walkFn(rel, info, err)
	})
}
