package dockerurl_test

import (
	"context"
	"testing"

	"github.com/google/go-github/v32/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thepwagner/action-update-go/updater"
)

func TestUpdater_Check(t *testing.T) {
	u := updaterFromFixture(t, "simple")

	cases := []struct {
		dep  updater.Dependency
		next *string
	}{
		{
			dep: updater.Dependency{
				Path:    "github.com/containerd/containerd/releases",
				Version: "v1.4.0",
			},
			next: github.String("v1.4.1"),
		},
		{
			dep: updater.Dependency{
				Path:    "github.com/hashicorp/terraform/releases",
				Version: "v0.13.0",
			},
			next: github.String("v0.13.3"),
		},
	}

	for _, tc := range cases {
		t.Run(tc.dep.Path, func(t *testing.T) {
			upd, err := u.Check(context.Background(), tc.dep)
			require.NoError(t, err)

			if tc.next == nil {
				assert.Nil(t, upd)
			} else if assert.NotNil(t, upd, "no update") {
				assert.Equal(t, *tc.next, upd.Next)
			}
		})
	}
}
