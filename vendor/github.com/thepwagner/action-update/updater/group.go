package updater

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/semver"
)

type Group struct {
	// Identify the group and members:
	// Name labels unique groups
	Name string `yaml:"name"`
	// Pattern is a prefix for the dependency, or a regular expression enclosed by /'s
	Pattern string `yaml:"pattern"`

	// Parameters that apply to members:
	// Range is a comma separated list of allowed semver ranges
	Range     string    `yaml:"range"`
	Frequency Frequency `yaml:"frequency"`

	compiledPattern *regexp.Regexp
}

type Frequency string

const (
	FrequencyDaily  Frequency = "daily"
	FrequencyWeekly Frequency = "weekly"
)

func (g *Group) Validate() error {
	if g.Name == "" {
		return fmt.Errorf("groups must specify name")
	}
	if g.Pattern == "" {
		return fmt.Errorf("groups must specify pattern")
	}
	switch g.Frequency {
	case "", FrequencyDaily, FrequencyWeekly:
	default:
		return fmt.Errorf("frequency must be: [%s,%s]", FrequencyDaily, FrequencyWeekly)
	}

	if strings.HasPrefix(g.Pattern, "/") && strings.HasSuffix(g.Pattern, "/") {
		re, err := regexp.Compile(g.Pattern[1 : len(g.Pattern)-1])
		if err != nil {
			return fmt.Errorf("compiling pattern: %w", err)
		}
		g.compiledPattern = re
	} else {
		g.compiledPattern = regexp.MustCompile("^" + regexp.QuoteMeta(g.Pattern))
	}
	return nil
}

func (g Group) InRange(v string) bool {
	for _, rangeCond := range strings.Split(g.Range, ",") {
		rangeCond = strings.TrimSpace(rangeCond)
		switch {
		case strings.HasPrefix(rangeCond, "<="):
			if semver.Compare(cleanRange(rangeCond, 2), v) < 0 {
				return false
			}
		case strings.HasPrefix(rangeCond, ">="):
			if semver.Compare(cleanRange(rangeCond, 2), v) > 0 {
				return false
			}
		case strings.HasPrefix(rangeCond, "<"):
			if semver.Compare(cleanRange(rangeCond, 1), v) <= 0 {
				return false
			}
		case strings.HasPrefix(rangeCond, ">"):
			if semver.Compare(cleanRange(rangeCond, 1), v) >= 0 {
				return false
			}
		}
	}
	return true
}

func cleanRange(rangeCond string, prefixLen int) string {
	s := strings.TrimSpace(rangeCond[prefixLen:])
	if !strings.HasPrefix(s, "v") {
		return fmt.Sprintf("v%s", s)
	}
	return s
}
