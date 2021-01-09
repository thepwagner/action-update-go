package updateaction

import (
	"bufio"
	"crypto/sha512"
	"fmt"
	"strings"

	"github.com/bmatcuk/doublestar/v3"
	"github.com/thepwagner/action-update/actions"
)

// Environment extends actions.Environment with configuration specific to update actions.
type Environment struct {
	actions.Environment

	// Inputs common to every updater:
	GitHubToken            string `env:"INPUT_TOKEN"`
	InputSigningKey        string `env:"INPUT_SIGNING_KEY"`
	InputGroups            string `env:"INPUT_GROUPS"`
	InputBranches          string `env:"INPUT_BRANCHES"`
	NoPush                 bool   `env:"INPUT_NO_PUSH"`
	InputIgnore            string `env:"INPUT_IGNORE"`
	InputDispatchOnRelease string `env:"INPUT_DISPATCH_ON_RELEASE"`
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

func (e *Environment) SigningKey() []byte {
	h := sha512.New()
	_, _ = fmt.Fprint(h, e.InputSigningKey)
	return h.Sum(nil)
}

func (e *Environment) Ignored(path string) bool {
	s := bufio.NewScanner(strings.NewReader(e.InputIgnore))
	s.Split(bufio.ScanWords)
	for s.Scan() {
		f := s.Text()
		if m, _ := doublestar.Match(f, path); m {
			return true
		}
	}
	return false
}

func (e *Environment) ReleaseDispatchRepos() (repos []string) {
	for _, b := range strings.Split(e.InputDispatchOnRelease, "\n") {
		if s := strings.TrimSpace(b); s != "" {
			repos = append(repos, s)
		}
	}
	return
}

// UpdateEnvironment smuggles *Environment out of structs that embed one.
type UpdateEnvironment interface{ env() *Environment }

var _ UpdateEnvironment = (*Environment)(nil)

func (e *Environment) env() *Environment { return e }
