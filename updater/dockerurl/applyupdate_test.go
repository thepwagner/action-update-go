package dockerurl_test

import (
	"io/ioutil"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thepwagner/action-update-go/updater"
	"github.com/thepwagner/action-update-go/updatertest"
)

func TestUpdater_ApplyUpdate_Simple(t *testing.T) {
	tempDir := updatertest.ApplyUpdateToFixture(t, "simple", updaterFactory(), update)

	b, err := ioutil.ReadFile(filepath.Join(tempDir, "Dockerfile"))
	require.NoError(t, err)
	dockerfile := string(b)
	assert.NotContains(t, dockerfile, "CONTAINERD_VERSION=1.4.0")
	assert.Contains(t, dockerfile, "CONTAINERD_VERSION=1.4.1")

	// https://github.com/containerd/containerd/releases/download/v1.4.0/containerd-1.4.0-linux-amd64.tar.gz.sha256sum
	assert.NotContains(t, dockerfile, "CONTAINERD_SHASUM=1401ff0b102f15f499598ceeb95f10ee37fa13a7c7ab57a2c764472665d55860")
	assert.Contains(t, dockerfile, "CONTAINERD_SHASUM=744ca14c62f5b2189c9fc22e9353508e49c85ab16038f3d74a852e6842e34fa3")
}

func TestUpdater_ApplyUpdate_Hash(t *testing.T) {
	t.Skip("fetches zipfile")

	elixirUpdate := updater.Update{
		Path:     "github.com/elixir-lang/elixir/releases",
		Previous: "v1.10.3",
		Next:     "v1.10.4",
	}
	tempDir := updatertest.ApplyUpdateToFixture(t, "hash", updaterFactory(), elixirUpdate)

	b, err := ioutil.ReadFile(filepath.Join(tempDir, "Dockerfile"))
	require.NoError(t, err)
	dockerfile := string(b)
	assert.NotContains(t, dockerfile, "ELIXIR_VERSION=v1.10.3")
	assert.Contains(t, dockerfile, "ELIXIR_VERSION=v1.10.4")

	assert.NotContains(t, dockerfile, "ELIXIR_CHECKSUM=fc6d06ad4cc596b2b6e4f01712f718200c69f3b9c49c7d3b787f9a67b36482658490cf01109b0b0842fc9d88a27f64a9aba817231498d99fa01fa99688263d55")
	assert.Contains(t, dockerfile, "ELIXIR_CHECKSUM=9727ae96d187d8b64e471ff0bb5694fcd1009cdcfd8b91a6b78b7542bb71fca59869d8440bb66a2523a6fec025f1d23394e7578674b942274c52b44e19ba2d43")
}
