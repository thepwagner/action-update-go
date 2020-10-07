package updater

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// Groups is an ordered list of Group with unique names.
// Prefer a list with .Name to map[string]Group for clear iteration order.
type Groups []*Group

func ParseGroups(s string) (Groups, error) {
	g := Groups{}
	if err := yaml.Unmarshal([]byte(s), &g); err != nil {
		return nil, err
	}
	if err := g.Validate(); err != nil {
		return nil, err
	}
	return g, nil
}

func (g Groups) Validate() error {
	uniqNames := map[string]struct{}{}
	for _, group := range g {
		if err := group.Validate(); err != nil {
			return fmt.Errorf("invalid group %q: %w", group.Name, err)
		}
		if _, ok := uniqNames[group.Name]; ok {
			return fmt.Errorf("duplicate group name: %q", group.Name)
		}
		uniqNames[group.Name] = struct{}{}
	}
	return nil
}

func (g Groups) ByName(name string) *Group {
	for _, group := range g {
		if group.Name == name {
			return group
		}
	}
	return nil
}

// GroupDependencies groups dependencies according to this configuration.
func (g Groups) GroupDependencies(deps []Dependency) (byGroupName map[string][]Dependency, ungrouped []Dependency) {
	byGroupName = make(map[string][]Dependency, len(g))
	for _, dep := range deps {
		group := g.matchGroup(dep)
		if group != "" {
			byGroupName[group] = append(byGroupName[group], dep)
		} else {
			ungrouped = append(ungrouped, dep)
		}
	}
	return
}

func (g Groups) matchGroup(dep Dependency) string {
	for _, group := range g {
		if group.compiledPattern.MatchString(dep.Path) {
			return group.Name
		}
	}
	return ""
}
