package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	"github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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

	env, err := cloneAndEnv(ctx, target, dir)
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

func cloneAndEnv(ctx context.Context, target, dir string) (*cmd.Environment, error) {
	targetURL, err := url.Parse(target)
	if err != nil {
		return nil, fmt.Errorf("parsing target URL: %w", err)
	}
	if targetURL.Host != "github.com" {
		return nil, fmt.Errorf("unsupported host")
	}

	if err := viper.WriteConfig(); err != nil {
		return nil, err
	}
	gh := gitrepo.NewGitHubClient(viper.GetString(flagGitHubToken))

	// Identify the repo from the provided URL, initialize remote:
	pathParts := strings.Split(targetURL.Path, "/")
	if len(pathParts) < 3 {
		return nil, fmt.Errorf("could not parse repo from path")
	}
	owner := pathParts[1]
	repo := pathParts[2]
	if err := exec.CommandExecute(ctx, dir, "git", "init", "."); err != nil {
		return nil, fmt.Errorf("git init: %w", err)
	}
	remoteURL := fmt.Sprintf("https://github.com/%s/%s", owner, repo)
	if err := exec.CommandExecute(ctx, dir, "git", "remote", "add", gitrepo.RemoteName, remoteURL); err != nil {
		return nil, fmt.Errorf("git remote add: %w", err)
	}

	// TODO: find references in .github/workflows/*.yaml to interpret configuration?

	// Interpret the path to decide how to clone and what event to simulate:
	if len(pathParts) == 3 {
		// owner+repo - fetch the default branch and simulate a "schedule" event
		remoteRef := path.Join("refs/remotes/origin", branchName)
		refSpec := fmt.Sprintf("+:%s", remoteRef)
		if err := exec.CommandExecute(ctx, dir, "git", "-c", "protocol.version=2", "fetch", "--no-tags",
			"--prune", "--progress", "--no-recurse-submodules", "--depth=1", gitrepo.RemoteName, refSpec); err != nil {
			return nil, fmt.Errorf("git fetch: %w", err)
		}
		if err := exec.CommandExecute(ctx, dir, "git", "checkout", "--progress", "--force", "-B", branchName, remoteRef); err != nil {
			return nil, fmt.Errorf("git fetch: %w", err)
		}

		return &cmd.Environment{
			GitHubEventName: "schedule",
		}, nil
	}

	if len(pathParts) == 5 && pathParts[3] == "pull" {
		// pull request - fetch the pr HEAD and simulate a "reopened" event
		prNumber, err := strconv.Atoi(pathParts[4])
		if err != nil {
			return nil, fmt.Errorf("parsing PR number: %w", err)
		}

		pr, _, err := gh.PullRequests.Get(ctx, owner, repo, prNumber)
		if err != nil {
			return nil, fmt.Errorf("getting PR: %w", err)
		}

		remoteRef := fmt.Sprintf("refs/remotes/pull/%d/merge", prNumber)
		refSpec := fmt.Sprintf("+%s:%s", pr.GetMergeCommitSHA(), remoteRef)
		if err := exec.CommandExecute(ctx, dir, "git", "-c", "protocol.version=2", "fetch", "--no-tags",
			"--prune", "--progress", "--no-recurse-submodules", "--depth=1", gitrepo.RemoteName, refSpec); err != nil {
			return nil, fmt.Errorf("git fetch: %w", err)
		}
		if err := exec.CommandExecute(ctx, dir, "git", "checkout", "--progress", "--force", remoteRef); err != nil {
			return nil, fmt.Errorf("git fetch: %w", err)
		}

		tmpEvt, err := tmpEventFile(&github.PullRequestEvent{
			Action:      github.String("reopened"),
			PullRequest: pr,
		})
		return &cmd.Environment{
			GitHubEventName: "pull_request",
			GitHubEventPath: tmpEvt,
		}, nil
	}

	return nil, nil
}

func tmpEventFile(evt interface{}) (string, error) {
	f, err := ioutil.TempFile("", "action-update-go-event-*")
	if err != nil {
		return "", err
	}
	defer f.Close()

	if err := json.NewEncoder(f).Encode(evt); err != nil {
		return "", err
	}
	return f.Name(), nil
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
