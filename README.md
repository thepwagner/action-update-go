# action-update-go

This action checks for available dependency updates to a go project, and opens individual pull requests proposing each available update.

* Ignores dependencies not released with semver
* Go module major version updates (e.g. `github.com/foo/bar/v2`)
* Vendoring detection and support
* Can multiple multiple base branches
* Update batching

Suggested triggers: `schedule`, `workflow_dispatch`.


## Simplest setup

```yaml
steps:
- uses: actions/checkout@v2
  # If you use Actions "push" for CI too, a Personal Access Token is required for update PRs to trigger
  with:
    token: ${{ secrets.MY_GITHUB_PAT }}
- uses: actions/setup-go@v2
  with:
    go-version: '1.15.0'  # or whatever version you use
- uses: thepwagner/action-update-go@main
  # If you use Actions "pull_request" for CI too, a Personal Access Token is required for update PRs to trigger
  with:
    token: ${{ secrets.MY_GITHUB_PAT }}
```

## Private dependencies

If your project has dependencies that require authentication, you can configure before invoking the action:

```yaml
- uses: actions/checkout@v2
- uses: actions/setup-go@v2
  with:
    go-version: '1.15.0'
- run: git config --global url."https://x-access-token:${GITHUB_TOKEN}@github.com".insteadOf "https://github.com"
  env:
    GITHUB_TOKEN: ${{ secrets.MY_GITHUB_PAT }}
- uses: thepwagner/action-update-go@main
  with:
    token: ${{ secrets.MY_GITHUB_PAT }}
```


#### But wait, there's more!

This also contains some alternative updater implementations, in various states of quality.
These should eventually be standalone Actions, but APIs are so fresh/churning that doing so would mean disrepair.

This is temporarily a multi-module repo, to be forked to N repositories (1 per module) before `v1`.

* `action-update` - shared code for updating dependencies from a GitHub Action
* `cli` - a CLI interface for testing, biased towards emulating Actions checkout for consistency.
* `docker` - incomplete updater for updating Dockerfile images (e.g. `FROM`)
* `dockerurl` - updater implementation for updating "GitHub Release" URLs in Dockerfile (e.g. `ENV`/`ARG`)
* `go` - updater implementation for go modules
