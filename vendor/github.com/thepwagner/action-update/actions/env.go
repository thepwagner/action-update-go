package actions

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/bmatcuk/doublestar/v2"
	"github.com/google/go-github/v32/github"
	"github.com/sirupsen/logrus"
)

// Environment includes Actions environment
// https://docs.github.com/en/free-pro-team@latest/actions/reference/environment-variables
type Environment struct {
	GitHubEventName  string `env:"GITHUB_EVENT_NAME"`
	GitHubEventPath  string `env:"GITHUB_EVENT_PATH"`
	GitHubRepository string `env:"GITHUB_REPOSITORY"`

	InputLogLevel string `env:"INPUT_LOG_LEVEL" envDefault:"info"`
	InputIgnore   string `env:"INPUT_IGNORE"`
}

// ParseEvent returns deserialized GitHub webhook payload, or an error.
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

// LogLevel returns the logrus level
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

func (e *Environment) Ignored(path string) bool {
	for _, p := range strings.Split(e.InputIgnore, "\n") {
		p = strings.TrimSpace(p)
		if m, _ := doublestar.Match(p, path); m {
			return true
		}
	}
	return false
}

// ActionEnvironment smuggles *Environment out of structs that embed one.
type ActionEnvironment interface{ env() *Environment }

var _ ActionEnvironment = (*Environment)(nil)

func (e *Environment) env() *Environment { return e }
