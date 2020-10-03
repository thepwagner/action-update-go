package cmd

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

type Config struct {
	// Actions environment:
	// https://docs.github.com/en/free-pro-team@latest/actions/reference/environment-variables
	GitHubEventName  string `env:"GITHUB_EVENT_NAME"`
	GitHubEventPath  string `env:"GITHUB_EVENT_PATH"`
	GitHubRepository string `env:"GITHUB_REPOSITORY"`

	// Inputs common to every updater:
	GitHubToken     string `env:"INPUT_TOKEN"`
	InputSigningKey []byte `env:"INPUT_SIGNING_KEY"`
	InputBatches    string `env:"INPUT_BATCHES"`
	InputBranches   string `env:"INPUT_BRANCHES"`
	InputLogLevel   string `env:"INPUT_LOG_LEVEL" envDefault:"debug"`
	NoPush          bool   `env:"INPUT_NO_PUSH"`
}

// ParseEvent returns deserialized GitHub webhook payload, or an error.
func (e *Config) ParseEvent() (interface{}, error) {
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

// Branches returns slice of all configured branches to update.
func (e *Config) Branches() (branches []string) {
	for _, b := range strings.Split(e.InputBranches, "\n") {
		if s := strings.TrimSpace(b); s != "" {
			branches = append(branches, s)
		}
	}
	return
}

// Batches returns a simple update batching configuration
func (e *Config) Batches() (map[string][]string, error) {
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

// LogLevel returns the logrus level
func (e *Config) LogLevel() logrus.Level {
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

type cfg interface{ cfg() *Config }

func (e *Config) cfg() *Config { return e }
