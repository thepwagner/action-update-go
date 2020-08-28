package gh

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/go-github/v31/github"
	"github.com/shurcooL/githubv4"
	"github.com/sirupsen/logrus"
	"golang.org/x/oauth2"
)

type Client struct {
	v3 *github.Client
	v4 *githubv4.Client
}

// FIXME: be an integration

func NewStaticTokenClient(ctx context.Context, token string) *Client {
	src := oauth2.StaticTokenSource(
		&oauth2.Token{AccessToken: token},
	)
	httpClient := oauth2.NewClient(ctx, src)
	httpClient.Transport = &headerHack{og: httpClient.Transport}

	return &Client{
		v3: github.NewClient(httpClient),
		v4: githubv4.NewClient(httpClient),
	}
}

type headerHack struct {
	og http.RoundTripper
}

func (h *headerHack) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Accept", "application/vnd.github.packages-preview+json")
	return h.og.RoundTrip(req)
}

func (c *Client) PackageFiles(ctx context.Context, owner, repo, version string) (map[string]string, error) {
	logrus.WithFields(logrus.Fields{
		"owner":   owner,
		"repo":    repo,
		"version": version,
	}).Debug("fetching package files")

	var q struct {
		Repository struct {
			PackageOwner struct {
				Packages struct {
					Edges []struct {
						Node struct {
							Name        string
							PackageType string
							Version     struct {
								Files struct {
									Edges []struct {
										Node struct {
											Name string
											URL  string
										}
									}
								} `graphql:"files(first: 10)"`
							} `graphql:"version(version: $version)"`
						}
					}
				} `graphql:"packages(first: 5)"`
			} `graphql:"... on PackageOwner"`
			Description string
		} `graphql:"repository(owner: $owner, name: $name)"`
	}

	variables := map[string]interface{}{
		"owner":   githubv4.String(owner),
		"name":    githubv4.String(repo),
		"version": githubv4.String(version),
	}
	if err := c.v4.Query(ctx, &q, variables); err != nil {
		return nil, err
	}

	ret := make(map[string]string)
	for _, pkg := range q.Repository.PackageOwner.Packages.Edges {
		for _, f := range pkg.Node.Version.Files.Edges {
			ret[f.Node.Name] = f.Node.URL
		}
	}

	return ret, nil
}

var releaseYAML = `
name: Release spike
on: push

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v2
      with:
        ref: %s
    - name: Set up JDK 1.8
      uses: actions/setup-java@v1
      with:
        java-version: 1.8
    - run: cat build.gradle.kts | tr '\n' '@' | sed -e's/publishing {@.*@}@//g' | tr '@' '\n' | tee gradle_base.txt
    - run: |
        cat <<EOF | tee gradle_publish.txt
        publishing {
          publications {
            create<MavenPublication>("maven") {
              groupId = "com.github.%s"
              artifactId = "%s"
              version = "%s"
              from(components["java"])
            }
          }
          repositories {
            maven {
              name = "GitHubPackages"
              url = uri("https://maven.pkg.github.com/%s/%s")
              credentials {
                username = "%s"
                password = System.getenv("GITHUB_TOKEN")
              }
            }
          }
        }
        EOF
    - run: cat gradle_base.txt gradle_publish.txt > build.gradle.kts
    - run: ./gradlew publishMavenPublicationToGitHubPackagesRepository
      env:
        GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
`

func (c *Client) TriggerPackageBuild(ctx context.Context, owner, repo, version string) error {
	branch := fmt.Sprintf("jithub/%s", version)
	baseRef, _, err := c.v3.Git.GetRef(ctx, owner, repo, "refs/heads/master")
	if err != nil {
		return err
	}

	branchRef := fmt.Sprintf("refs/heads/%s", branch)
	_, _, err = c.v3.Git.CreateRef(ctx, owner, repo, &github.Reference{
		Ref:    &branchRef,
		Object: baseRef.Object,
	})
	if err != nil {
		return err
	}

	workflowYAML := fmt.Sprintf(releaseYAML, version, owner, repo, version, owner, repo, owner)

	message := fmt.Sprintf(":shipit: %s", version)
	_, _, err = c.v3.Repositories.CreateFile(ctx, owner, repo, ".github/workflows/release.yaml", &github.RepositoryContentFileOptions{
		Branch:  &branch,
		Message: &message,
		Content: []byte(workflowYAML),
	})
	return err
}
