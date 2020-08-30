package gitrepo_test

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"sync/atomic"
	"testing"

	"github.com/go-git/go-git/v5/plumbing"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thepwagner/action-update-go/gomod"
)

func TestSharedSandbox_ReadFile(t *testing.T) {
	sbx := initSharedSandbox(t)
	defer sbx.Close()

	b, err := sbx.ReadFile(fileName)
	require.NoError(t, err)
	assert.Equal(t, fileData, b)
}

func TestSharedSandbox_WriteFile(t *testing.T) {
	for _, fn := range []string{fileName, "not-" + fileName} {
		t.Run(fn, func(t *testing.T) {
			sbx := initSharedSandbox(t)
			defer sbx.Close()

			randomBytes := make([]byte, 10)
			_, err := rand.Read(randomBytes)
			require.NoError(t, err)
			err = sbx.WriteFile(fn, randomBytes)
			require.NoError(t, err)

			b, err := sbx.ReadFile(fn)
			require.NoError(t, err)
			assert.Equal(t, randomBytes, b)
		})
	}
}

func TestSharedSandbox_Cmd(t *testing.T) {
	sbx := initSharedSandbox(t)
	defer sbx.Close()

	data := "data"
	err := sbx.Cmd(context.Background(), "bash", "-c", fmt.Sprintf("echo -n %q > %s", data, fileName))
	require.NoError(t, err)

	b, err := sbx.ReadFile(fileName)
	require.NoError(t, err)
	assert.Equal(t, data, string(b))
}

func TestSharedSandbox_Walk(t *testing.T) {
	sbx := initSharedSandbox(t)
	defer sbx.Close()

	var ctr int64
	err := sbx.Walk(func(path string, info os.FileInfo, err error) error {
		assert.Equal(t, fileName, path)
		atomic.AddInt64(&ctr, 1)
		return nil
	})
	require.NoError(t, err)
	assert.Equal(t, int64(1), ctr)
}

func initSharedSandbox(t *testing.T) gomod.Sandbox {
	sharedRepo := initSingleTree(t, plumbing.NewBranchReferenceName(branchName))
	sbx, err := sharedRepo.NewSandbox(branchName, "my-awesome-branch")
	require.NoError(t, err)
	return sbx
}
