package docker_test

import (
	"strings"
	"testing"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thepwagner/action-update-docker/docker"
)

func TestInterpolation_Interpolate(t *testing.T) {
	parsed, err := parser.Parse(strings.NewReader(`FROM scratch
ENV FOO=bar FOZ=baz
ENV FOOBAR 123
ARG NESTED=SUPER${FOO}AND$FOZTOO
`))
	require.NoError(t, err)

	i := docker.NewInterpolation(parsed)
	cases := map[string]string{
		"$FOO":       "bar",
		"$FOOBAR":    "123",
		"${FOO}":     "bar",
		"${FOO}BAR":  "barBAR",
		"$FOO$FOZ":   "barbaz",
		"$FOZZYBEAR": "bazZYBEAR",
		"$FOX":       "$FOX", // what is your sound?
		"$NESTED":    "SUPERbarANDbazTOO",
	}

	for s, expected := range cases {
		t.Run(s, func(t *testing.T) {
			assert.Equal(t, expected, i.Interpolate(s))
		})
	}
}
