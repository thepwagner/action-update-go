module github.com/thepwagner/action-update-cli

require (
	github.com/google/go-github/v32 v32.1.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.7.1
	github.com/thepwagner/action-update v1.0.0
	github.com/thepwagner/action-update-docker v1.0.0
	github.com/thepwagner/action-update-dockerurl v0.0.1
	github.com/thepwagner/action-update-go v0.0.1
)

replace (
	github.com/containerd/containerd => github.com/containerd/containerd v1.4.0
	github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200916142827-bd33bbf0497b+incompatible
	github.com/thepwagner/action-update => ../action-update
	github.com/thepwagner/action-update-docker => ../docker
	github.com/thepwagner/action-update-dockerurl => ../dockerurl
	github.com/thepwagner/action-update-go => ../go
)

go 1.15
