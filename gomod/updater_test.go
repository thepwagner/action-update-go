package gomod_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/otiai10/copy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thepwagner/action-update-go/gomod"
)

var goModFiles = []string{gomod.GoModFn, gomod.GoSumFn}

func TestUpdater_ApplyUpdate_Simple(t *testing.T) {
	tempDir := updateFixture(t, &gomod.Updater{Tidy: true}, "simple", gomod.Update{
		Path: "github.com/pkg/errors",
		Next: "v0.8.1",
	})

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
	tempDir := updateFixture(t, &gomod.Updater{Tidy: false}, "simple", gomod.Update{
		Path: "github.com/pkg/errors",
		Next: "v0.8.1",
	})

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
	tempDir := updateFixture(t, &gomod.Updater{Tidy: true}, "vendor", gomod.Update{
		Path: "github.com/pkg/errors",
		Next: "v0.8.1",
	})

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
	tempDir := updateFixture(t, &gomod.Updater{Tidy: true}, "major", gomod.Update{
		Path:     "github.com/caarlos0/env/v5",
		Previous: "v5.1.4",
		Next:     "v6.2.0",
	})

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
	tempDir := updateFixture(t, &gomod.Updater{Tidy: true}, "notinroot", gomod.Update{
		Path: "github.com/pkg/errors",
		Next: "v0.8.1",
	})

	// Path is renamed in module files:
	for _, fn := range goModFiles {
		b, err := ioutil.ReadFile(filepath.Join(tempDir, fn))
		require.NoError(t, err)
		s := string(b)

		assert.NotContains(t, s, "github.com/pkg/errors v0.8.0")
		assert.Contains(t, s, "github.com/pkg/errors v0.8.1")
	}

	var files []string
	_ = filepath.Walk(tempDir, func(path string, fi os.FileInfo, err error) error {
		if fi.IsDir() {
			return nil
		}
		rel, _ := filepath.Rel(tempDir, path)
		files = append(files, rel)
		return nil
	})
	sort.Strings(files)

	assert.Equal(t, []string{"cmd/main.go", "go.mod", "go.sum"}, files)
}

func updateFixture(t *testing.T, updater *gomod.Updater, fixture string, update gomod.Update) string {
	tempDir := t.TempDir()
	err := copy.Copy(fmt.Sprintf("../fixtures/%s", fixture), tempDir)
	require.NoError(t, err)

	err = updater.ApplyUpdate(context.Background(), tempDir, update)
	require.NoError(t, err)
	return tempDir
}
