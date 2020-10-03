package updater_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	updater2 "github.com/thepwagner/action-update/updater"
)

const (
	baseBranch = "main"
	mockPath   = "github.com/foo/bar"
)

var mockUpdate = updater2.Update{
	Path: mockPath,
	Next: "v2.0.0",
}

func TestUpdatesByBranch(t *testing.T) {
	byBranch := updater2.UpdatesByBranch{}
	assert.Len(t, byBranch, 0)
}

func TestUpdatesByBranch_AddOpen(t *testing.T) {
	byBranch := updater2.UpdatesByBranch{}
	byBranch.AddOpen(baseBranch, mockUpdate)

	assert.Equal(t, updater2.UpdatesByBranch{
		baseBranch: {
			Open: []updater2.Update{mockUpdate},
		},
	}, byBranch)

	byBranch.AddOpen(baseBranch, mockUpdate)
	assert.Equal(t, updater2.UpdatesByBranch{
		baseBranch: {
			Open: []updater2.Update{mockUpdate},
		},
	}, byBranch)
}

func TestUpdatesByBranch_AddClosed(t *testing.T) {
	byBranch := updater2.UpdatesByBranch{}
	byBranch.AddClosed(baseBranch, mockUpdate)

	assert.Equal(t, updater2.UpdatesByBranch{
		baseBranch: {
			Closed: []updater2.Update{mockUpdate},
		},
	}, byBranch)
}

func TestUpdates_OpenUpdate(t *testing.T) {
	u := updater2.Updates{Open: []updater2.Update{mockUpdate}}

	cases := map[string]struct {
		Update updater2.Update
		Open   string
	}{
		"open": {
			Update: mockUpdate,
			Open:   mockUpdate.Next,
		},
		"higher version": {
			Update: updater2.Update{Path: mockPath, Next: "v3.0.0"},
			Open:   mockUpdate.Next,
		},
		"lower version": {
			Update: updater2.Update{Path: mockPath, Next: "v1.0.0"},
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
	u := updater2.Updates{Open: []updater2.Update{mockUpdate}}
	assert.Equal(t, "", u.ClosedUpdate(mockUpdate))
}
