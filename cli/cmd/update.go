package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"net/url"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var keep bool

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Perform dependency updates",
	RunE: func(cmd *cobra.Command, args []string) error {
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
	dirLog := logrus.WithField("tempDir", dir)
	dirLog.Debug("created tempdir")
	if !keep {
		defer func() {
			if err := os.RemoveAll(dir); err != nil {
				dirLog.WithError(err).Warn("error cleaning temp dir")
			}
		}()
	}

	if err := gitClone(ctx, target, dir); err != nil {
		return err
	}

	return nil
}

func gitClone(ctx context.Context, target, dir string) error {
	targetURL, err := url.Parse(target)
	if err != nil {
		return fmt.Errorf("parsing target URL: %w", err)
	}
	if targetURL.Host != "github.com" {
		return fmt.Errorf("unsupported host")
	}
	pathParts := strings.Split(targetURL.Path, "/")
	if len(pathParts) <= 2 {
		owner := pathParts[0]
		repo := pathParts[1]
		err := exec.Comm
	}

	return nil
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
