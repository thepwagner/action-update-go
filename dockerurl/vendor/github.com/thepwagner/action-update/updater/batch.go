package updater

import (
	"sort"
	"strings"
)

func GroupDependencies(batches map[string][]string, deps []Dependency) map[string][]Dependency {
	// Build an inverted index of prefixes, and list of all prefixes sorted by length:
	index := map[string]string{}
	var prefixes []string
	for name, prefixList := range batches {
		prefixes = append(prefixes, prefixList...)
		for _, p := range prefixList {
			index[p] = name
		}
	}
	sort.Slice(prefixes, func(i, j int) bool {
		return len(prefixes[i]) > len(prefixes[j])
	})

	// Sort dependencies into groups:
	groups := make(map[string][]Dependency, len(batches))
	for _, dep := range deps {
		// Find the closest prefix, there may not be one:
		var branchName string
		for _, p := range prefixes {
			if strings.HasPrefix(dep.Path, p) {
				branchName = index[p]
				break
			}
		}

		groups[branchName] = append(groups[branchName], dep)
	}

	return groups
}
