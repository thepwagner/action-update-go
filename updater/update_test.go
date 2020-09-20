package updater_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thepwagner/action-update-go/updater"
)

const (
	baseBranch = "main"
	mockPath   = "github.com/foo/bar"
)

var mockUpdate = updater.Update{
	Path: mockPath,
	Next: "v2.0.0",
}

func TestUpdatesByBranch(t *testing.T) {
	byBranch := updater.UpdatesByBranch{}
	assert.Len(t, byBranch, 0)
}

func TestUpdatesByBranch_AddOpen(t *testing.T) {
	byBranch := updater.UpdatesByBranch{}
	byBranch.AddOpen(baseBranch, mockUpdate)

	assert.Equal(t, updater.UpdatesByBranch{
		baseBranch: {
			Open: []updater.Update{mockUpdate},
		},
	}, byBranch)

	byBranch.AddOpen(baseBranch, mockUpdate)
	assert.Equal(t, updater.UpdatesByBranch{
		baseBranch: {
			Open: []updater.Update{mockUpdate},
		},
	}, byBranch)
}

func TestUpdatesByBranch_AddClosed(t *testing.T) {
	byBranch := updater.UpdatesByBranch{}
	byBranch.AddClosed(baseBranch, mockUpdate)

	assert.Equal(t, updater.UpdatesByBranch{
		baseBranch: {
			Closed: []updater.Update{mockUpdate},
		},
	}, byBranch)
}

func TestUpdates_OpenUpdate(t *testing.T) {
	u := updater.Updates{Open: []updater.Update{mockUpdate}}

	cases := map[string]struct {
		Update updater.Update
		Open   string
	}{
		"open": {
			Update: mockUpdate,
			Open:   mockUpdate.Next,
		},
		"higher version": {
			Update: updater.Update{Path: mockPath, Next: "v3.0.0"},
			Open:   mockUpdate.Next,
		},
		"lower version": {
			Update: updater.Update{Path: mockPath, Next: "v1.0.0"},
			Open:   "",
		},
	}

	for label, tc := range cases {
		t.Run(label, func(t *testing.T) {
			assert.Equal(t, tc.Open, u.OpenUpdate(tc.Update))
		})
	}
}

func TestUpdates_ClosedUpdate(t *testing.T) {
	u := updater.Updates{Open: []updater.Update{mockUpdate}}
	assert.Equal(t, "", u.ClosedUpdate(mockUpdate))
}
