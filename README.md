# action-update-go

This action checks for available dependency updates to a go project, and opens individual pull requests proposing each available update.

* Ignores dependencies not released with semver
* Ignores dependencies if the initial PR is closed
* Go module major version updates (e.g. `github.com/foo/bar/v2`)
* Vendoring detection and support
* Can multiple multiple base branches

Suggested triggers: `schedule`, `workflow_dispatch`.


## Simplest setup

```yaml
steps:
- uses: actions/checkout@v2
- uses: actions/setup-go@v2
  with:
    go-version: '1.15.0'  # or whatever version you use
- uses: thepwagner/action-update-go@main
  with:
    # If you use Actions for CI too, a Personal Access Token is required for update PRs to trigger
    token: ${{ secrets.MY_GITHUB_PAT }}
```

## Private dependencies

If your project depends on dependencies that require authentication, you can configure before invoking the action:

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
