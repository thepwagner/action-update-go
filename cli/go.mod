module github.com/thepwagner/action-update-cli

require (
	github.com/google/go-github/v32 v32.1.0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/cobra v1.0.0
	github.com/spf13/viper v1.7.1
	github.com/thepwagner/action-update v0.0.1
	github.com/thepwagner/action-update-go v0.0.1
)

replace (
	github.com/thepwagner/action-update => ../action-update
	github.com/thepwagner/action-update-go => ../go
)

go 1.15
