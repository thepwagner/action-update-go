package docker

import (
	"fmt"
	"sort"
	"strings"

	"github.com/moby/buildkit/frontend/dockerfile/command"
	"github.com/moby/buildkit/frontend/dockerfile/parser"
)

// Interpolation attempts to interpolate a variable Dockerfile string
// Easily fooled by duplicate ARGs
type Interpolation struct {
	vars      map[string]string
	varsByLen []string
}

func NewInterpolation(parsed *parser.Result) *Interpolation {
	i := &Interpolation{
		vars: map[string]string{},
	}
	for _, instruction := range parsed.AST.Children {
		switch instruction.Value {
		case command.Arg:
			varSplit := strings.SplitN(instruction.Next.Value, "=", 2)
			if len(varSplit) == 2 {
				i.vars[varSplit[0]] = varSplit[1]
			}

		case command.Env:
			iter := instruction
			for iter.Next != nil {
				i.vars[iter.Next.Value] = iter.Next.Next.Value
				iter = iter.Next.Next
			}
		}
	}

	i.varsByLen = make([]string, 0, len(i.vars))
	for k := range i.vars {
		i.varsByLen = append(i.varsByLen, k)
	}
	sort.SliceStable(i.varsByLen, func(x, y int) bool {
		return len(i.varsByLen[x]) > len(i.varsByLen[y])
	})

	return i
}

func (i *Interpolation) Interpolate(s string) string {
	pre := s
	for _, k := range i.varsByLen {
		v := i.vars[k]
		s = strings.ReplaceAll(s, fmt.Sprintf("${%s}", k), v)
		s = strings.ReplaceAll(s, fmt.Sprintf("$%s", k), v)
	}
	if pre != s {
		return i.Interpolate(s)
	}
	return s
}
