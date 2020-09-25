package gomod_test

import (
	"context"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thepwagner/action-update-go/gomod"
	"github.com/thepwagner/action-update-go/updater"
)

var goModFiles = []string{gomod.GoModFn, gomod.GoSumFn}

var pkgErrors081 = updater.Update{
	Path: "github.com/pkg/errors",
	Next: "v0.8.1",
}

func TestUpdater_ApplyUpdate_Simple(t *testing.T) {
	tempDir := applyUpdateToFixture(t, "simple", pkgErrors081)
	for _, fn := range goModFiles {
		b, err := ioutil.ReadFile(filepath.Join(tempDir, fn))
		require.NoError(t, err)
		s := string(b)

		// Target path is updated:
		assert.NotContains(t, s, "github.com/pkg/errors v0.8.0")
		assert.Contains(t, s, "github.com/pkg/errors v0.8.1")

		// Other (outdated) paths are unchanged:
		assert.Contains(t, s, "github.com/sirupsen/logrus v1.5.0")
	}
}

func TestUpdater_ApplyUpdate_Simple_NoTidy(t *testing.T) {
	tempDir := applyUpdateToFixture(t, "simple", pkgErrors081, gomod.WithTidy(false))

	b, err := ioutil.ReadFile(filepath.Join(tempDir, gomod.GoModFn))
	require.NoError(t, err)
	goMod := string(b)
	assert.NotContains(t, goMod, "github.com/pkg/errors v0.8.0")
	assert.Contains(t, goMod, "github.com/pkg/errors v0.8.1")

	b, err = ioutil.ReadFile(filepath.Join(tempDir, gomod.GoSumFn))
	require.NoError(t, err)
	goSum := string(b)
	assert.Contains(t, goSum, "github.com/pkg/errors v0.8.0")
	assert.Contains(t, goSum, "github.com/pkg/errors v0.8.1")
}

func TestUpdater_ApplyUpdate_Vendor(t *testing.T) {
	tempDir := applyUpdateToFixture(t, "vendor", pkgErrors081)

	b, err := ioutil.ReadFile(filepath.Join(tempDir, gomod.VendorModulesFn))
	require.NoError(t, err)
	modulesTxt := string(b)
	assert.NotContains(t, modulesTxt, "github.com/pkg/errors v0.8.0")
	assert.Contains(t, modulesTxt, "github.com/pkg/errors v0.8.1")

	// Sample to verify the source code matches:
	// https://github.com/pkg/errors/compare/v0.8.0...v0.8.1
	b, err = ioutil.ReadFile(filepath.Join(tempDir, "vendor", "github.com", "pkg", "errors", "stack.go"))
	require.NoError(t, err)
	stackGo := string(b)
	assert.Contains(t, stackGo, `Format formats the stack of Frames according to the fmt.Formatter interface.`)
	assert.NotContains(t, stackGo, `segments from the beginning of the file path until the number of path`)
}

func TestUpdater_ApplyUpdate_Major(t *testing.T) {
	env6 := updater.Update{
		Path:     "github.com/caarlos0/env/v5",
		Previous: "v5.1.4",
		Next:     "v6.2.0",
	}
	tempDir := applyUpdateToFixture(t, "major", env6, gomod.WithMajorVersions(true))

	// Path is renamed in module files:
	for _, fn := range goModFiles {
		b, err := ioutil.ReadFile(filepath.Join(tempDir, fn))
		require.NoError(t, err)
		s := string(b)

		assert.NotContains(t, s, "github.com/caarlos0/env/v5 v5.1.4")
		assert.Contains(t, s, "github.com/caarlos0/env/v6 v6.2.0")
	}

	// Path is updated in source code:
	b, err := ioutil.ReadFile(filepath.Join(tempDir, "main.go"))
	require.NoError(t, err)
	mainGo := string(b)
	assert.NotContains(t, mainGo, "github.com/caarlos0/env/v5")
	assert.Contains(t, mainGo, "github.com/caarlos0/env/v6")
}

func TestUpdater_ApplyUpdate_NotInRoot(t *testing.T) {
	tempDir := applyUpdateToFixture(t, "notinroot", pkgErrors081)

	// Path is renamed in module files:
	for _, fn := range goModFiles {
		b, err := ioutil.ReadFile(filepath.Join(tempDir, fn))
		require.NoError(t, err)
		s := string(b)

		assert.NotContains(t, s, "github.com/pkg/errors v0.8.0")
		assert.Contains(t, s, "github.com/pkg/errors v0.8.1")
	}

	// Walk and collect file paths:
	var files []string
	err := filepath.Walk(tempDir, func(path string, fi os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if fi.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(tempDir, path)
		files = append(files, rel)
		return nil
	})
	require.NoError(t, err)
	sort.Strings(files)

	assert.Equal(t, []string{"cmd/main.go", "go.mod", "go.sum"}, files, "update added/removed files")
}

func applyUpdateToFixture(t *testing.T, fixture string, up updater.Update, opts ...gomod.UpdaterOpt) string {
	tempDir := tempDirFromFixture(t, fixture)
	goUpdater := gomod.NewUpdater(tempDir, opts...)
	err := goUpdater.ApplyUpdate(context.Background(), up)
	require.NoError(t, err)
	return tempDir
}
