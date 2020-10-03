package repo_test

import (
	"os"
	"testing"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thepwagner/action-update/repo"
)

func TestNewGitHubRepo(t *testing.T) {
	gr := initGitRepo(t, plumbing.NewBranchReferenceName(branchName))

	gh, err := repo.NewGitHubRepo(gr, testKey, "foo/bar", "")
	require.NoError(t, err)
	assert.NotNil(t, gh)
}

func tokenOrSkip(t *testing.T) string {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		t.Skip("set GITHUB_TOKEN")
	}
	return token
}
