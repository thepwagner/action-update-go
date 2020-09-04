package repo_test

import (
	"context"
	"os"
	"testing"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thepwagner/action-update-go/repo"
)

func TestNewGitHubRepo(t *testing.T) {
	gr := initGitRepo(t, plumbing.NewBranchReferenceName(branchName))

	gh, err := repo.NewGitHubRepo(gr, "foo/bar", "")
	require.NoError(t, err)
	assert.NotNil(t, gh)
}

func TestGitHubRepo_Updates(t *testing.T) {
	token := tokenOrSkip(t)

	gr := initGitRepo(t, plumbing.NewBranchReferenceName(branchName))
	gh, err := repo.NewGitHubRepo(gr, "thepwagner/action-update-go", token)
	require.NoError(t, err)

	updates, err := gh.Updates(context.Background())
	require.NoError(t, err)
	assert.Len(t, updates, 3)

	mainUpdates := updates["main"]
	assert.Len(t, mainUpdates.Open, 0)
	assert.Len(t, mainUpdates.Closed, 4)
}

func tokenOrSkip(t *testing.T) string {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		t.Skip("set GITHUB_TOKEN")
	}
	return token
}
