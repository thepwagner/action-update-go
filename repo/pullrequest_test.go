package repo_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thepwagner/action-update-go/repo"
	"github.com/thepwagner/action-update-go/updater"
)

var (
	awsSdkGo13417 = updater.Update{
		Path:     "github.com/aws/aws-sdk-go",
		Previous: "v1.34.16",
		Next:     "v1.34.17",
	}
	fooBar987 = updater.Update{
		Path:     "github.com/foo/bar",
		Previous: "v0.4.1",
		Next:     "v99.88.77",
	}
)

func TestGitHubPullRequestContentGenerator_Generate(t *testing.T) {
	token := tokenOrSkip(t)
	client := repo.NewGitHubClient(token)
	gen := repo.NewGitHubPullRequestContent(client)

	title, body, err := gen.Generate(context.Background(), awsSdkGo13417)
	require.NoError(t, err)
	assert.Equal(t, "Update github.com/aws/aws-sdk-go from v1.34.16 to v1.34.17", title)
	assert.Equal(t, strings.TrimSpace(`
Here is github.com/aws/aws-sdk-go v1.34.17, I hope it works.

[changelog](https://github.com/aws/aws-sdk-go/blob/v1.34.17/CHANGELOG.md)

`+"```json"+`
{
  "major": false,
  "minor": false,
  "patch": true
}
`+"```"), strings.TrimSpace(body))
}

func TestGitHubPullRequestContentGenerator_GenerateNoChangeLog(t *testing.T) {
	token := tokenOrSkip(t)
	client := repo.NewGitHubClient(token)
	gen := repo.NewGitHubPullRequestContent(client)

	title, body, err := gen.Generate(context.Background(), fooBar987)
	require.NoError(t, err)
	assert.Equal(t, "Update github.com/foo/bar from v0.4.1 to v99.88.77", title)
	assert.Equal(t, strings.TrimSpace(`
Here is github.com/foo/bar v99.88.77, I hope it works.

`+"```json"+`
{
  "major": true,
  "minor": false,
  "patch": false
}
`+"```"), strings.TrimSpace(body))
}

func TestGitHubPullRequestContentGenerator_GenerateMultiple(t *testing.T) {
	token := tokenOrSkip(t)
	client := repo.NewGitHubClient(token)
	gen := repo.NewGitHubPullRequestContent(client)

	title, body, err := gen.Generate(context.Background(), awsSdkGo13417, fooBar987)
	require.NoError(t, err)
	assert.Equal(t, "Dependency Updates", title)
	assert.Equal(t, strings.TrimSpace(`
Here are some updates, I hope they work.

#### github.com/aws/aws-sdk-go@v1.34.17

[changelog](https://github.com/aws/aws-sdk-go/blob/v1.34.17/CHANGELOG.md)

#### github.com/foo/bar@v99.88.77

`+"```json"+`
{
  "major": true,
  "minor": false,
  "patch": false
}
`+"```"), strings.TrimSpace(body))
}
