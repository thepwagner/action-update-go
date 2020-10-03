package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/thepwagner/action-update-go/actions"
	"github.com/thepwagner/action-update-go/cmd"
	"github.com/thepwagner/action-update-go/common/exec"
	gitrepo "github.com/thepwagner/action-update-go/repo"
)

// TODO: wire to cobra
var (
	keep       bool
	branchName string = "master"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Perform dependency updates",
	RunE: func(cmd *cobra.Command, args []string) error {
		logrus.SetLevel(logrus.DebugLevel)
		var target string
		if len(args) > 0 {
			target = args[0]
		} else {
			target = "https://github.com/thepwagner/action-update-go"
		}
		return MockUpdate(context.Background(), target)
	},
}

func MockUpdate(ctx context.Context, target string) error {
	logrus.WithField("target", target).Info("performing mock update")

	// Setup a tempdir for the clone:
	dir, err := ioutil.TempDir("", "action-update-go-*")
	if err != nil {
		return err
	}
	dirLog := logrus.WithField("temp_dir", dir)
	dirLog.Debug("created tempdir")
	if !keep {
		defer func() {
			if err := os.RemoveAll(dir); err != nil {
				dirLog.WithError(err).Warn("error cleaning temp dir")
			}
		}()
	}

	env, err := cloneAndSetEnv(ctx, target, dir)
	if err != nil {
		return err
	} else if env == nil {
		return fmt.Errorf("could not detect environment")
	}
	dirLog.Info("cloned to tempdir")

	if err := os.Chdir(dir); err != nil {
		return err
	}

	return cmd.HandleEvent(ctx, env, actions.Handlers)
}

func cloneAndSetEnv(ctx context.Context, target, dir string) (*cmd.Environment, error) {
	targetURL, err := url.Parse(target)
	if err != nil {
		return nil, fmt.Errorf("parsing target URL: %w", err)
	}
	if targetURL.Host != "github.com" {
		return nil, fmt.Errorf("unsupported host")
	}

	// Interpret the path to decide how to clone, and set environment variables so
	pathParts := strings.Split(targetURL.Path, "/")
	if len(pathParts) <= 3 {
		owner := pathParts[1]
		repo := pathParts[2]
		if err := exec.CommandExecute(ctx, dir, "git", "init", "."); err != nil {
			return nil, fmt.Errorf("git init: %w", err)
		}
		remoteURL := fmt.Sprintf("https://github.com/%s/%s", owner, repo)
		if err := exec.CommandExecute(ctx, dir, "git", "remote", "add", gitrepo.RemoteName, remoteURL); err != nil {
			return nil, fmt.Errorf("git remote add: %w", err)
		}

		remoteRef := path.Join("refs/remotes/origin", branchName)
		refSpec := fmt.Sprintf("+:%s", remoteRef)
		if err := exec.CommandExecute(ctx, dir, "git", "-c", "protocol.version=2", "fetch",
			"--prune", "--progress", "--no-recurse-submodules", "--depth=1", gitrepo.RemoteName, refSpec); err != nil {
			return nil, fmt.Errorf("git fetch: %w", err)
		}

		if err := exec.CommandExecute(ctx, dir, "git", "checkout", "--progress", "--force", "-B", branchName, remoteRef); err != nil {
			return nil, fmt.Errorf("git fetch: %w", err)
		}
		return &cmd.Environment{
			GitHubEventName: "workflow_dispatch",
		}, nil
	}

	return nil, nil
}

func init() {
	rootCmd.AddCommand(updateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// updateCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// updateCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
