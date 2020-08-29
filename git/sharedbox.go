package gitrepo

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/sirupsen/logrus"
	"github.com/thepwagner/action-update-go/gomod"
)

type SharedSandbox struct {
	lock *sync.Mutex
	wt   *git.Worktree
}

func (s *SharedSandbox) ReadFile(path string) ([]byte, error) {
	return readWorktreeFile(s.wt, path)
}

func (s *SharedSandbox) WriteFile(path string, data []byte) error {
	out, err := s.wt.Filesystem.OpenFile(path, os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("opening go.mod: %w", err)
	}
	defer out.Close()
	if _, err := out.Write(data); err != nil {
		return fmt.Errorf("writing updated go.mod: %w", err)
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

func (s *SharedSandbox) Commit(ctx context.Context, message string, author gomod.GitIdentity) error {
	if err := s.commit(message, author); err != nil {
		return err
	}
	if err := s.push(ctx); err != nil {
		return err
	}
	return nil
}

func (s *SharedSandbox) Walk(walkFn filepath.WalkFunc) error {
	root := s.wt.Filesystem.Root()
	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		return walkFn(rel, info, err)
	})
}

func (s *SharedSandbox) commit(message string, author gomod.GitIdentity) error {
	when := time.Now()
	if err := worktreeAddAll(s.wt); err != nil {
		return err
	}

	commit, err := s.wt.Commit(message, &git.CommitOptions{
		Author: &object.Signature{
			Name:  author.Name,
			Email: author.Email,
			When:  when,
		},
		Committer: &object.Signature{
			Name:  author.Name,
			Email: author.Email,
			When:  when,
		},
	})
	if err != nil {
		return fmt.Errorf("committing branch: %w", err)
	}
	logrus.WithField("commit", commit.String()).Info("committed update")
	return nil
}

func worktreeAddAll(wt *git.Worktree) error {
	// wt.AddGlob() is attractive, but does not respect .gitignore
	// .Status() respects .gitignore so add file by file:
	status, err := wt.Status()
	if err != nil {
		return fmt.Errorf("checking status for add: %w", err)
	}
	for fn := range status {
		if _, err := wt.Add(fn); err != nil {
			return fmt.Errorf("adding file %q: %w", fn, err)
		}
	}

	logrus.WithField("files", len(status)).Debug("added files to index")
	return nil
}

func (s *SharedSandbox) push(ctx context.Context) error {
	// go-git supports Push, but not the [http "https://github.com/"] .gitconfig directive that actions/checkout uses for auth
	// we could extract from u.repo.Config().Raw, but who are we trying to impress?
	if err := s.Cmd(ctx, "git", "push", "-f"); err != nil {
		return fmt.Errorf("pushing: %w", err)
	}
	return nil
}

func (s *SharedSandbox) Close() error {
	s.lock.Unlock()
	return nil
}
