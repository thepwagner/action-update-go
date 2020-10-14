# action-update-go

## This is not endorsed by or associated with GitHub, Dependabot, etc.

This action checks for available dependency updates to a go project, and opens individual pull requests proposing each available update.

* Ignores dependencies not released with semver
* Go module major version updates (e.g. `github.com/foo/bar/v2`)
* Vendoring detection and support
* All the features common to [action-update](https://github.com/thepwagner/action-update) actions
  * Can monitor multiple base branches (e.g. `main`, `v1`)
  * Update batching
  * Post-update script hooks

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
