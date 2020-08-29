package gomod_test

import (
	"fmt"
	"testing"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/assert"
	"github.com/thepwagner/action-update-go/gomod"
)

func TestModuleUpdate_Major(t *testing.T) {
	cases := []struct {
		v1, v2   string
		expected bool
	}{
		{v1: "v1", v2: "v1", expected: false},
		{v1: "v1.0", v2: "v1.1", expected: false},
		{v1: "v1", v2: "v2", expected: true},
		{v1: "v2", v2: "v1", expected: true},
	}

	for _, c := range cases {
		t.Run(fmt.Sprintf("%s %s", c.v1, c.v2), func(t *testing.T) {
			u := gomod.ModuleUpdate{Previous: c.v1, Next: c.v2}
			assert.Equal(t, c.expected, u.Major())
		})
	}
}

func TestModuleUpdate_BranchName(t *testing.T) {
	u := gomod.ModuleUpdate{
		Path:     "github.com/stretchr/testify",
		Previous: "v1.6.0",
		Next:     "v1.6.1",
		Base:     plumbing.NewHashReference(plumbing.NewBranchReferenceName("main"), plumbing.ZeroHash),
	}
	assert.Equal(t, "action-update-go/main/github.com/stretchr/testify/v1.6.1", u.BranchName())
}
