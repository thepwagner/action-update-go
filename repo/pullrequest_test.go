package repo_test

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
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

func TestNewSignedUpdateDescriptor(t *testing.T) {
	key := []byte{1, 2, 3, 4}

	cases := []struct {
		signature string
		updates   []updater.Update
	}{
		{
			signature: "HAF6zSdBBOsbrLRClce7M73tN7VhCdPB6YYhECL/ifDC6DHR0YSGXoY6JQeEaFoncJbxp/afBpY+GVE5DUfWwQ==",
			updates:   []updater.Update{awsSdkGo13417},
		},
		{
			signature: "kq9CbO3rYkThJPiJgVTfhkfAG4q5aEeXuta0x3wPVdUnqQhitA/FasfJ2WftpfiZvueCnknoX04yxTM94BUn4A==",
			updates:   []updater.Update{fooBar987},
		},
		{
			signature: "TL6d3v5DKRu8uDY5doooDLLd7mJDHx6U5P4jRZYanLT4VI1dzt1gIRvZGW3G0ZlDQqmuTOftovTlwLHO1VW4Xw==",
			updates:   []updater.Update{awsSdkGo13417, fooBar987},
		},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("%v", tc.updates), func(t *testing.T) {
			descriptor, err := repo.NewSignedUpdateDescriptor(key, tc.updates...)
			require.NoError(t, err)

			buf, err := json.Marshal(&descriptor)
			require.NoError(t, err)
			t.Log(string(buf))

			assert.Equal(t, tc.updates, descriptor.Updates)
			assert.Equal(t, tc.signature, base64.StdEncoding.EncodeToString(descriptor.Signature))

			verified, err := repo.VerifySignedUpdateDescriptor(key, descriptor)
			require.NoError(t, err)
			assert.Equal(t, tc.updates, verified)
		})
	}
}

func TestVerifySignedUpdateDescriptor_Invalid(t *testing.T) {
	_, err := repo.VerifySignedUpdateDescriptor([]byte{}, repo.SignedUpdateDescriptor{})
	assert.EqualError(t, err, "invalid signature")
}
