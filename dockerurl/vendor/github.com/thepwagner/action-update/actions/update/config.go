package update

import (
	"fmt"
	"strings"

	"github.com/thepwagner/action-update/actions"
	"gopkg.in/yaml.v3"
)

type Config struct {
	actions.Config

	// Inputs common to every updater:
	GitHubToken     string `env:"INPUT_TOKEN"`
	InputSigningKey []byte `env:"INPUT_SIGNING_KEY"`
	InputBatches    string `env:"INPUT_BATCHES"`
	InputBranches   string `env:"INPUT_BRANCHES"`
	NoPush          bool   `env:"INPUT_NO_PUSH"`
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
