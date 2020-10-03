package repo_test

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thepwagner/action-update/repo"
	updater2 "github.com/thepwagner/action-update/updater"
)

var (
	awsSdkGo13417 = updater2.Update{
		Path:     "github.com/aws/aws-sdk-go",
		Previous: "v1.34.16",
		Next:     "v1.34.17",
	}
	fooBar987 = updater2.Update{
		Path:     "github.com/foo/bar",
		Previous: "v0.4.1",
		Next:     "v99.88.77",
	}
	testKey = []byte{1, 2, 3, 4}
)

func TestGitHubPullRequestContent_Generate(t *testing.T) {
	token := tokenOrSkip(t)
	client := repo.NewGitHubClient(token)
	gen := repo.NewGitHubPullRequestContent(client, testKey)

	title, body, err := gen.Generate(context.Background(), awsSdkGo13417)
	require.NoError(t, err)
	assert.Equal(t, "Update github.com/aws/aws-sdk-go from v1.34.16 to v1.34.17", title)
	assert.Equal(t, strings.TrimSpace(`
Here is github.com/aws/aws-sdk-go v1.34.17, I hope it works.

[changelog](https://github.com/aws/aws-sdk-go/blob/v1.34.17/CHANGELOG.md)

<!--::action-update-go::
{"updates":[{"path":"github.com/aws/aws-sdk-go","previous":"v1.34.16","next":"v1.34.17"}],"signature":"HAF6zSdBBOsbrLRClce7M73tN7VhCdPB6YYhECL/ifDC6DHR0YSGXoY6JQeEaFoncJbxp/afBpY+GVE5DUfWwQ=="}
-->
`), strings.TrimSpace(body))
}

func TestGitHubPullRequestContent_ParseBody(t *testing.T) {
	token := tokenOrSkip(t)
	client := repo.NewGitHubClient(token)
	gen := repo.NewGitHubPullRequestContent(client, testKey)

	body := `
<!--::action-update-go::
{"updates":[{"path":"github.com/aws/aws-sdk-go","previous":"v1.34.16","next":"v1.34.17"}],"signature":"HAF6zSdBBOsbrLRClce7M73tN7VhCdPB6YYhECL/ifDC6DHR0YSGXoY6JQeEaFoncJbxp/afBpY+GVE5DUfWwQ=="}
-->`
	parsed := gen.ParseBody(body)
	assert.Equal(t, []updater2.Update{awsSdkGo13417}, parsed)
}

func TestGitHubPullRequestContent_GenerateNoChangeLog(t *testing.T) {
	token := tokenOrSkip(t)
	client := repo.NewGitHubClient(token)
	gen := repo.NewGitHubPullRequestContent(client, testKey)

	title, body, err := gen.Generate(context.Background(), fooBar987)
	require.NoError(t, err)
	assert.Equal(t, "Update github.com/foo/bar from v0.4.1 to v99.88.77", title)
	assert.Equal(t, strings.TrimSpace(`
Here is github.com/foo/bar v99.88.77, I hope it works.

<!--::action-update-go::
{"updates":[{"path":"github.com/foo/bar","previous":"v0.4.1","next":"v99.88.77"}],"signature":"kq9CbO3rYkThJPiJgVTfhkfAG4q5aEeXuta0x3wPVdUnqQhitA/FasfJ2WftpfiZvueCnknoX04yxTM94BUn4A=="}
-->
`), strings.TrimSpace(body))
}

func TestGitHubPullRequestContent_GenerateMultiple(t *testing.T) {
	token := tokenOrSkip(t)
	client := repo.NewGitHubClient(token)
	gen := repo.NewGitHubPullRequestContent(client, testKey)

	title, body, err := gen.Generate(context.Background(), awsSdkGo13417, fooBar987)
	require.NoError(t, err)
	assert.Equal(t, "Dependency Updates", title)
	assert.Equal(t, strings.TrimSpace(`
Here are some updates, I hope they work.

#### github.com/aws/aws-sdk-go@v1.34.17

[changelog](https://github.com/aws/aws-sdk-go/blob/v1.34.17/CHANGELOG.md)

#### github.com/foo/bar@v99.88.77

<!--::action-update-go::
{"updates":[{"path":"github.com/aws/aws-sdk-go","previous":"v1.34.16","next":"v1.34.17"},{"path":"github.com/foo/bar","previous":"v0.4.1","next":"v99.88.77"}],"signature":"TL6d3v5DKRu8uDY5doooDLLd7mJDHx6U5P4jRZYanLT4VI1dzt1gIRvZGW3G0ZlDQqmuTOftovTlwLHO1VW4Xw=="}
-->
`), strings.TrimSpace(body))
}
