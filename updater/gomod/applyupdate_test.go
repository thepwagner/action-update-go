package gomod_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thepwagner/action-update-go/updater"
	"github.com/thepwagner/action-update-go/updater/gomod"
	"github.com/thepwagner/action-update-go/updatertest"
)

var pkgErrors081 = updater.Update{
	Path: "github.com/pkg/errors",
	Next: "v0.8.1",
}

func TestUpdater_ApplyUpdate_Simple(t *testing.T) {
	tempDir := updatertest.ApplyUpdateToFixture(t, "simple", updaterFactory(), pkgErrors081)
	uf := readModFiles(t, tempDir)

	for _, s := range uf.GoModFiles() {
		// Target path is updated:
		assert.NotContains(t, s, "github.com/pkg/errors v0.8.0")
		assert.Contains(t, s, "github.com/pkg/errors v0.8.1")

		// Other (outdated) paths are unchanged:
		assert.Contains(t, s, "github.com/sirupsen/logrus v1.5.0")
	}
}

func TestUpdater_ApplyUpdate_Simple_NoTidy(t *testing.T) {
	tempDir := updatertest.ApplyUpdateToFixture(t, "simple", updaterFactory(gomod.WithTidy(false)), pkgErrors081)
	uf := readModFiles(t, tempDir)

	assert.NotContains(t, uf.GoMod, "github.com/pkg/errors v0.8.0")
	assert.Contains(t, uf.GoMod, "github.com/pkg/errors v0.8.1")

	assert.Contains(t, uf.GoSum, "github.com/pkg/errors v0.8.0")
	assert.Contains(t, uf.GoSum, "github.com/pkg/errors v0.8.1")
}

func TestUpdater_ApplyUpdate_Vendor(t *testing.T) {
	tempDir := updatertest.ApplyUpdateToFixture(t, "vendor", updaterFactory(), pkgErrors081)
	uf := readModFiles(t, tempDir)

	assert.NotContains(t, uf.ModulesTxt, "github.com/pkg/errors v0.8.0")
	assert.Contains(t, uf.ModulesTxt, "github.com/pkg/errors v0.8.1")

	// Sample to verify the source code matches:
	// https://github.com/pkg/errors/compare/v0.8.0...v0.8.1
	b, err := ioutil.ReadFile(filepath.Join(tempDir, "vendor", "github.com", "pkg", "errors", "stack.go"))
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
	tempDir := updatertest.ApplyUpdateToFixture(t, "major", updaterFactory(gomod.WithMajorVersions(true)), env6)
	uf := readModFiles(t, tempDir)

	// Path is renamed in module files:
	for _, s := range uf.GoModFiles() {
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

func TestUpdater_ApplyUpdate_Major_Gopkg(t *testing.T) {
	yaml1 := updater.Update{
		Path:     "gopkg.in/yaml.v1",
		Previous: "v1.0.0",
		Next:     "v2.3.0",
	}
	tempDir := updatertest.ApplyUpdateToFixture(t, "gopkg", updaterFactory(gomod.WithMajorVersions(true)), yaml1)
	uf := readModFiles(t, tempDir)

	// Path is renamed in module files:
	for _, s := range uf.GoModFiles() {
		assert.NotContains(t, s, "gopkg.in/yaml.v1 v1.0.0")
		assert.Contains(t, s, "gopkg.in/yaml.v2 v2.3.0")
	}

	// Path is updated in source code:
	b, err := ioutil.ReadFile(filepath.Join(tempDir, "main.go"))
	require.NoError(t, err)
	mainGo := string(b)
	assert.NotContains(t, mainGo, "gopkg.in/yaml.v1")
	assert.Contains(t, mainGo, "gopkg.in/yaml.v2")
}

func TestUpdater_ApplyUpdate_NotInRoot(t *testing.T) {
	tempDir := updatertest.ApplyUpdateToFixture(t, "notinroot", updaterFactory(), pkgErrors081)
	uf := readModFiles(t, tempDir)

	// Path is renamed in module files:
	for _, s := range uf.GoModFiles() {
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

func TestUpdater_ApplyUpdate_Replace(t *testing.T) {
	replacement := updater.Update{
		Path: "github.com/thepwagner/errors",
		Next: "v0.8.1",
	}
	tempDir := updatertest.ApplyUpdateToFixture(t, "replace", updaterFactory(), replacement)
	uf := readModFiles(t, tempDir)

	// Base version unchanged:
	assert.Contains(t, uf.GoMod, "github.com/pkg/errors v0.8.0")
	// Replacement changed:
	assert.NotContains(t, uf.GoMod, "github.com/thepwagner/errors v0.8.0")
	assert.Contains(t, uf.GoMod, "github.com/thepwagner/errors v0.8.1")

	assert.NotContains(t, uf.GoSum, "github.com/pkg/errors")
	assert.NotContains(t, uf.GoSum, "github.com/thepwagner/errors v0.8.0")
	assert.Contains(t, uf.GoSum, "github.com/thepwagner/errors v0.8.1")
}

func TestUpdater_ApplyUpdate_MultimoduleCommon(t *testing.T) {
	logrus160 := updater.Update{
		Path: "github.com/sirupsen/logrus",
		Next: "v1.6.0",
	}

	tempDir := updatertest.ApplyUpdateToFixture(t, "multimodule/common", updaterFactory(), logrus160)
	uf := readModFiles(t, filepath.Join(tempDir, "common"))
	for _, s := range uf.GoModFiles() {
		assert.NotContains(t, s, "github.com/sirupsen/logrus v1.5.0")
		assert.Contains(t, s, "github.com/sirupsen/logrus v1.6.0")
	}
}

func TestUpdater_ApplyUpdate_MultimoduleCmd(t *testing.T) {
	tempDir := updatertest.ApplyUpdateToFixture(t, "multimodule/cmd", updaterFactory(), pkgErrors081)
	uf := readModFiles(t, filepath.Join(tempDir, "cmd"))
	for _, s := range uf.GoModFiles() {
		assert.NotContains(t, s, "github.com/pkg/errors v0.8.0")
		assert.Contains(t, s, "github.com/pkg/errors v0.8.1")
	}
}

func TestUpdater_ApplyUpdate_Multimodule(t *testing.T) {
	tempDir := updatertest.ApplyUpdateToFixture(t, "multimodule", updaterFactory(), pkgErrors081)
	uf := readModFiles(t, filepath.Join(tempDir, "cmd"))
	for _, s := range uf.GoModFiles() {
		assert.NotContains(t, s, "github.com/pkg/errors v0.8.0")
		assert.Contains(t, s, "github.com/pkg/errors v0.8.1")
	}
}

type modFiles struct {
	GoMod, GoSum string
	ModulesTxt   string
}

func (uf modFiles) GoModFiles() []string {
	return []string{uf.GoMod, uf.GoSum}
}

func readModFiles(t *testing.T, tempDir string) (uf modFiles) {
	b, err := ioutil.ReadFile(filepath.Join(tempDir, gomod.GoModFn))
	require.NoError(t, err)
	uf.GoMod = string(b)

	b, err = ioutil.ReadFile(filepath.Join(tempDir, gomod.GoSumFn))
	require.NoError(t, err)
	uf.GoSum = string(b)

	b, err = ioutil.ReadFile(filepath.Join(tempDir, gomod.VendorModulesFn))
	if err == nil {
		uf.ModulesTxt = string(b)
	}
	return
}
