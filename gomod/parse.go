package gomod

import (
	"fmt"
	"io/ioutil"

	"github.com/dependabot/gomodules-extracted/cmd/go/_internal_/modfile"
)

const goModFn = "go.mod"

// Parse returns parsed `go.mod`.
func Parse() (*modfile.File, error) {
	b, err := ioutil.ReadFile(goModFn)
	if err != nil {
		return nil, fmt.Errorf("reading go.mod: %w", err)
	}
	parsed, err := modfile.Parse(goModFn, b, nil)
	if err != nil {
		return nil, fmt.Errorf("parsing go.mod: %w", err)
	}
	return parsed, nil
}
