module github.com/thepwagner/action-update-dockerurl

go 1.15

require (
	github.com/sirupsen/logrus v1.7.0
	github.com/stretchr/testify v1.6.1
	github.com/thepwagner/action-update v1.0.0
	github.com/thepwagner/action-update-docker v1.0.0
)

replace (
	github.com/thepwagner/action-update => ../action-update
	github.com/thepwagner/action-update-docker => ../docker
)
