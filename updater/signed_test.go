package updater_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/thepwagner/action-update-go/updater"
)

var (
	testKey = []byte{1, 2, 3, 4}
)

func TestNewSignedUpdateDescriptor(t *testing.T) {
	var (
		update1 = updater.Update{Path: "github.com/foo/bar", Previous: "v1.0.0", Next: "v1.1.0"}
		update2 = updater.Update{Path: "github.com/foo/baz", Previous: "v1.0.0", Next: "v2.0.0"}
	)

	cases := []struct {
		signature string
		updates   []updater.Update
	}{
		{
			signature: "TCQ5Vfa1pOKaMIhVpxVOWS/EzTuq+5AFfwrO8cuKxhes/hJ6xDrusp2YtFz2Vbc+pOYyu5oLbQBnyc9REk5mfA==",
			updates:   []updater.Update{update1},
		},
		{
			signature: "2H4Ka0Yzk5GyGRGusbZMLYdi+7a+EHJVmArdFLgFLBVzqNTDdnimHNbHym5v38h/lO8f2sObzVQPewa3TiytFw==",
			updates:   []updater.Update{update1, update2},
		},
		{
			signature: "2H4Ka0Yzk5GyGRGusbZMLYdi+7a+EHJVmArdFLgFLBVzqNTDdnimHNbHym5v38h/lO8f2sObzVQPewa3TiytFw==",
			updates:   []updater.Update{update2, update1},
		},
	}

	for _, tc := range cases {
		t.Run(fmt.Sprintf("%v", tc.updates), func(t *testing.T) {
			descriptor, err := updater.NewSignedUpdateDescriptor(testKey, tc.updates...)
			require.NoError(t, err)

			buf, err := json.Marshal(&descriptor)
			require.NoError(t, err)
			t.Log(string(buf))

			assert.Equal(t, tc.updates, descriptor.Updates)
			assert.Equal(t, tc.signature, base64.StdEncoding.EncodeToString(descriptor.Signature))

			verified, err := updater.VerifySignedUpdateDescriptor(testKey, descriptor)
			require.NoError(t, err)
			assert.Equal(t, tc.updates, verified)
		})
	}
}

func TestVerifySignedUpdateDescriptor_Invalid(t *testing.T) {
	_, err := updater.VerifySignedUpdateDescriptor([]byte{}, updater.SignedUpdateDescriptor{})
	assert.EqualError(t, err, "invalid signature")
}
