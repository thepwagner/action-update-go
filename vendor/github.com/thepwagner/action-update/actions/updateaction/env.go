package updateaction

import (
	"strings"

	"github.com/thepwagner/action-update/actions"
)

// Environment extends actions.Environment with configuration specific to update actions.
type Environment struct {
	actions.Environment

	// Inputs common to every updater:
	GitHubToken     string `env:"INPUT_TOKEN"`
	InputSigningKey []byte `env:"INPUT_SIGNING_KEY"`
	InputGroups     string `env:"INPUT_GROUPS"`
	InputBranches   string `env:"INPUT_BRANCHES"`
	NoPush          bool   `env:"INPUT_NO_PUSH"`
}

// Branches returns slice of all configured branches to update.
func (e *Environment) Branches() (branches []string) {
	for _, b := range strings.Split(e.InputBranches, "\n") {
		if s := strings.TrimSpace(b); s != "" {
			branches = append(branches, s)
		}
	}
	return
}

// UpdateEnvironment smuggles *Environment out of structs that embed one.
type UpdateEnvironment interface{ env() *Environment }

var _ UpdateEnvironment = (*Environment)(nil)

func (e *Environment) env() *Environment { return e }
