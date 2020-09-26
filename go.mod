module github.com/thepwagner/action-update-go

go 1.15

require (
	github.com/caarlos0/env/v5 v5.1.4
	github.com/dependabot/gomodules-extracted v1.1.0
	github.com/go-git/go-git/v5 v5.1.0
	github.com/google/go-cmp v0.5.2 // indirect
	github.com/google/go-github/v32 v32.1.0
	github.com/moby/buildkit v0.7.2
	github.com/otiai10/copy v1.2.0
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.6.0
	golang.org/x/oauth2 v0.0.0-20200107190931-bf48bf16ab8d
	gopkg.in/yaml.v2 v2.3.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c
)

replace github.com/containerd/containerd => github.com/containerd/containerd v1.4.0

replace github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200916142827-bd33bbf0497b+incompatible
