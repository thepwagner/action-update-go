module github.com/thepwagner/action-update-docker

go 1.15

require (
	github.com/dependabot/gomodules-extracted v1.1.0
	github.com/google/go-cmp v0.5.2 // indirect
	github.com/moby/buildkit v0.7.2
	github.com/stretchr/testify v1.6.1
	github.com/thepwagner/action-update v0.0.1
)

replace (
	github.com/containerd/containerd => github.com/containerd/containerd v1.4.0
	github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200916142827-bd33bbf0497b+incompatible
)
