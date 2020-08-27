package cmd

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/caarlos0/env/v6"
	"github.com/google/go-github/v32/github"
)

type Environment struct {
	GitHubEventName  string `env:"GITHUB_EVENT_NAME"`
	GitHubEventPath  string `env:"GITHUB_EVENT_PATH"`
	InputBranches    string `env:"INPUT_BRANCHES"`
	GitHubRepository string `env:"GITHUB_REPOSITORY"`
	GitHubToken      string `env:"GITHUB_TOKEN"`
}

func ParseEnvironment() (Environment, error) {
	var e Environment
	if err := env.Parse(&e); err != nil {
		return Environment{}, fmt.Errorf("parsing environment: %w", err)
	}
	return e, nil
}

func (e Environment) ParseEvent() (interface{}, error) {
	switch e.GitHubEventName {
	case "schedule", "workflow_dispatch":
		return nil, nil
	}

	var evt interface{}
	b, err := ioutil.ReadFile(e.GitHubEventPath)
	if err != nil {
		return nil, fmt.Errorf("reading event %q: %w", e.GitHubEventPath, err)
	}

	evt, err = github.ParseWebHook(e.GitHubEventName, b)
	if err != nil {
		return nil, fmt.Errorf("parsing event: %w", err)
	}
	return evt, nil
}

func (e Environment) Branches() (branches []string) {
	for _, b := range strings.Split(e.InputBranches, "\n") {
		if s := strings.TrimSpace(b); s != "" {
			branches = append(branches, s)
		}
	}
	return
}
