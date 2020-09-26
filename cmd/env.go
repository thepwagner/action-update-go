package cmd

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/caarlos0/env/v5"
	"github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type Environment struct {
	GitHubEventName  string `env:"GITHUB_EVENT_NAME"`
	GitHubEventPath  string `env:"GITHUB_EVENT_PATH"`
	GitHubRepository string `env:"GITHUB_REPOSITORY"`

	InputBatches  string `env:"INPUT_BATCHES"`
	InputBranches string `env:"INPUT_BRANCHES"`
	GitHubToken   string `env:"INPUT_TOKEN"`
	InputLogLevel string `env:"INPUT_LOG_LEVEL" envDefault:"debug"`
}

func ParseEnvironment() (*Environment, error) {
	var e Environment
	if err := env.Parse(&e); err != nil {
		return nil, fmt.Errorf("parsing environment: %w", err)
	}
	return &e, nil
}

func (e *Environment) ParseEvent() (interface{}, error) {
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

func (e *Environment) Branches() (branches []string) {
	for _, b := range strings.Split(e.InputBranches, "\n") {
		if s := strings.TrimSpace(b); s != "" {
			branches = append(branches, s)
		}
	}
	return
}

func (e *Environment) Batches() (map[string][]string, error) {
	raw := map[string]interface{}{}
	if err := yaml.Unmarshal([]byte(e.InputBatches), &raw); err != nil {
		return nil, fmt.Errorf("decoding batches yaml: %w", err)
	}

	m := make(map[string][]string, len(raw))
	for key, value := range raw {
		var prefixes []string
		switch v := value.(type) {
		case []interface{}:
			for _, s := range v {
				prefixes = append(prefixes, fmt.Sprintf("%v", s))
			}
		case string:
			prefixes = append(prefixes, v)
		}
		m[key] = prefixes
	}
	return m, nil
}

func (e *Environment) LogLevel() logrus.Level {
	if e.InputLogLevel == "" {
		return logrus.InfoLevel
	}

	lvl, err := logrus.ParseLevel(e.InputLogLevel)
	if err != nil {
		logrus.WithError(err).Warn("could not parse log level")
		lvl = logrus.InfoLevel
	}
	return lvl
}
