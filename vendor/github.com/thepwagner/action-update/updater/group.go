package updater

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/mod/semver"
)

type Group struct {
	// Identify the group and members:
	// Name labels unique groups
	Name string `yaml:"name"`
	// Pattern is a prefix for the dependency, or a regular expression enclosed by /'s
	Pattern string `yaml:"pattern"`

	// Parameters that apply to members:
	// Range is a comma separated list of allowed semver ranges
	Range      string `yaml:"range"`
	CoolDown   string `yaml:"cooldown"`
	PreScript  string `yaml:"pre-script"`
	PostScript string `yaml:"post-script"`

	compiledPattern *regexp.Regexp
}

func (g *Group) Validate() error {
	if g.Name == "" {
		return fmt.Errorf("groups must specify name")
	}
	if g.Pattern == "" {
		return fmt.Errorf("groups must specify pattern")
	}
	if !durPattern.MatchString(g.CoolDown) {
		return fmt.Errorf("invalid cooldown, expected ISO8601 duration: %q", g.CoolDown)
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

func (g *Group) InRange(v string) bool {
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

var durPattern = regexp.MustCompile(`P?(((?P<year>\d+)Y)?((?P<month>\d+)M)?((?P<day>\d+)D)|(?P<week>\d+)W)?`)

const (
	oneYear  = 8766 * time.Hour
	oneMonth = 730*time.Hour + 30*time.Minute
	oneWeek  = 7 * 24 * time.Hour
	oneDay   = 24 * time.Hour
)

func (g *Group) CoolDownDuration() time.Duration {
	m := durPattern.FindStringSubmatch(g.CoolDown)

	var ret time.Duration
	for i, name := range durPattern.SubexpNames() {
		part := m[i]
		if i == 0 || name == "" || part == "" {
			continue
		}

		val, err := strconv.Atoi(part)
		if err != nil {
			return 0
		}
		valDur := time.Duration(val)
		switch name {
		case "year":
			ret += valDur * oneYear
		case "month":
			ret += valDur * oneMonth
		case "week":
			ret += valDur * oneWeek
		case "day":
			ret += valDur * oneDay
		}
	}

	return ret
}
