module github.com/thepwagner/action-update-go/multimodule/cmd

require (
	github.com/pkg/errors v0.8.0
	github.com/thepwagner/action-update-go/multimodule/common v1.0.0
)

replace github.com/thepwagner/action-update-go/multimodule/common => ../common

go 1.15
