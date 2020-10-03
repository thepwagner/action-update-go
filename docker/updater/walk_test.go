package updater_test

import (
	"fmt"
	"sync/atomic"
	"testing"

	"github.com/moby/buildkit/frontend/dockerfile/parser"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thepwagner/action-update-go/updater/docker"
	updater2 "github.com/thepwagner/action-update/updater"
)

const fixtureCount = 2

func TestWalkDockerfiles(t *testing.T) {
	var cnt int64
	deps, err := docker.ExtractDockerfileDependencies("testdata/", func(_ *parser.Result) ([]updater2.Dependency, error) {
		cur := atomic.AddInt64(&cnt, 1)
		return []updater2.Dependency{{Path: "test", Version: fmt.Sprintf("v%d", cur)}}, nil
	})
	require.NoError(t, err)

	assert.Equal(t, int64(fixtureCount), cnt, "function not invoked N times")
	assert.Len(t, deps, fixtureCount, "walk did not collect N results")
}
