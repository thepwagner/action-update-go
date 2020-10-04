package actions

import (
	"fmt"
	"io/ioutil"

	"github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"
)

// Config includes Actions environment
// https://docs.github.com/en/free-pro-team@latest/actions/reference/environment-variables
type Config struct {
	GitHubEventName  string `env:"GITHUB_EVENT_NAME"`
	GitHubEventPath  string `env:"GITHUB_EVENT_PATH"`
	GitHubRepository string `env:"GITHUB_REPOSITORY"`

	InputLogLevel string `env:"INPUT_LOG_LEVEL" envDefault:"info"`
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

type actionsConfig interface{ cfg() *Config }

var _ actionsConfig = (*Config)(nil)

func (e *Config) cfg() *Config { return e }
