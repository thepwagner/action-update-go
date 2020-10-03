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
	gitrepo "github.com/thepwagner/action-update-go/repo"
)

const (
	flagKeepTmpDir = "Keep"
	flagBranchName = "Branch"
)

var updateCmd = &cobra.Command{
	Use:   "update <url>",
	Short: "Perform dependency updates",
	RunE: func(cmd *cobra.Command, args []string) error {
		viper.SetDefault(flagKeepTmpDir, false)
		viper.SetDefault(flagBranchName, "master")

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
	if !viper.GetBool(flagKeepTmpDir) {
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
	env.NoPush = true

	if err := os.Chdir(dir); err != nil {
		return err
	}

	return cmd.HandleEvent(ctx, env, actions.Handlers)
}

func cloneAndEnv(ctx context.Context, target, dir string) (*cmd.Environment, error) {
	parsed, err := parseTargetURL(target)
	if err != nil {
		return nil, err
	}

	if err := parsed.initRepo(ctx, dir); err != nil {
		return nil, err
	}

	ghToken := viper.GetString(flagGitHubToken)
	gh := gitrepo.NewGitHubClient(ghToken)

	// TODO: find self-references in .github/workflows/*.yaml to guess configuration?

	// Interpret the path to decide how to clone and what event to simulate:
	if parsed.prNumber > 0 {
		// Pull request - fetch PR head and simulate `pull_request.reopened` to recreate
		return parsed.clonePullRequest(ctx, gh, dir)
	}

	// Fetch default branch and simulate `schedule` to reopen all
	return parsed.cloneEvent(ctx, dir)
}

type parsedTarget struct {
	owner, repo string
	prNumber    int
}

func parseTargetURL(target string) (*parsedTarget, error) {
	targetURL, err := url.Parse(target)
	if err != nil {
		return nil, fmt.Errorf("parsing target URL: %w", err)
	}
	if targetURL.Host != "github.com" {
		return nil, fmt.Errorf("unsupported host")
	}
	pathParts := strings.Split(targetURL.Path, "/")
	if len(pathParts) < 3 {
		return nil, fmt.Errorf("could not parse repo from path")
	}

	t := &parsedTarget{
		owner: pathParts[1],
		repo:  pathParts[2],
	}
	if len(pathParts) == 5 && pathParts[3] == "pull" {
		t.prNumber, err = strconv.Atoi(pathParts[4])
		if err != nil {
			return nil, fmt.Errorf("parsing PR number: %w", err)
		}
	}
	return t, nil
}

func (p *parsedTarget) initRepo(ctx context.Context, dir string) error {
	if err := cmd.CommandExecute(ctx, dir, "git", "init", "."); err != nil {
		return fmt.Errorf("git init: %w", err)
	}
	remoteURL := fmt.Sprintf("https://github.com/%s/%s", p.owner, p.repo)
	if err := cmd.CommandExecute(ctx, dir, "git", "remote", "add", gitrepo.RemoteName, remoteURL); err != nil {
		return fmt.Errorf("git remote add: %w", err)
	}
	return nil
}

func (p *parsedTarget) clonePullRequest(ctx context.Context, gh *github.Client, dir string) (*cmd.Environment, error) {
	// pull request - fetch the pr HEAD and simulate a "reopened" event
	pr, _, err := gh.PullRequests.Get(ctx, p.owner, p.repo, p.prNumber)
	if err != nil {
		return nil, fmt.Errorf("getting PR: %w", err)
	}

	remoteRef := fmt.Sprintf("refs/remotes/pull/%d/merge", p.prNumber)
	refSpec := fmt.Sprintf("+%s:%s", pr.GetMergeCommitSHA(), remoteRef)
	if err := cmd.CommandExecute(ctx, dir, "git", "-c", "protocol.version=2", "fetch", "--no-tags",
		"--prune", "--progress", "--no-recurse-submodules", "--depth=1", gitrepo.RemoteName, refSpec); err != nil {
		return nil, fmt.Errorf("git fetch: %w", err)
	}
	if err := cmd.CommandExecute(ctx, dir, "git", "checkout", "--progress", "--force", remoteRef); err != nil {
		return nil, fmt.Errorf("git fetch: %w", err)
	}

	tmpEvt, err := tmpEventFile(&github.PullRequestEvent{
		Action:      github.String("reopened"),
		PullRequest: pr,
	})
	if err != nil {
		return nil, fmt.Errorf("creating temp event file: %w", err)
	}
	return &cmd.Environment{
		GitHubEventName: "pull_request",
		GitHubEventPath: tmpEvt,
	}, nil
}

func (p *parsedTarget) cloneEvent(ctx context.Context, dir string) (*cmd.Environment, error) {
	branchName := viper.GetString(flagBranchName)
	remoteRef := path.Join("refs/remotes/origin", branchName)
	refSpec := fmt.Sprintf("+:%s", remoteRef)
	if err := cmd.CommandExecute(ctx, dir, "git", "-c", "protocol.version=2", "fetch", "--no-tags",
		"--prune", "--progress", "--no-recurse-submodules", "--depth=1", gitrepo.RemoteName, refSpec); err != nil {
		return nil, fmt.Errorf("git fetch: %w", err)
	}
	if err := cmd.CommandExecute(ctx, dir, "git", "checkout", "--progress", "--force", "-B", branchName, remoteRef); err != nil {
		return nil, fmt.Errorf("git fetch: %w", err)
	}

	return &cmd.Environment{
		GitHubEventName: "schedule",
	}, nil
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
}
