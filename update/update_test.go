package update_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/thepwagner/action-update-go/update"
)

const (
	baseBranch = "main"
	mockPath   = "github.com/foo/bar"
)

var mockUpdate = update.Update{
	Path: mockPath,
	Next: "v2.0.0",
}

func TestUpdatesByBranch(t *testing.T) {
	byBranch := update.UpdatesByBranch{}
	assert.Len(t, byBranch, 0)
}

func TestUpdatesByBranch_AddOpen(t *testing.T) {
	byBranch := update.UpdatesByBranch{}
	byBranch.AddOpen(baseBranch, mockUpdate)

	assert.Equal(t, update.UpdatesByBranch{
		baseBranch: {
			Open: []update.Update{mockUpdate},
		},
	}, byBranch)

	byBranch.AddOpen(baseBranch, mockUpdate)
	assert.Equal(t, update.UpdatesByBranch{
		baseBranch: {
			Open: []update.Update{mockUpdate},
		},
	}, byBranch)
}

func TestUpdatesByBranch_AddClosed(t *testing.T) {
	byBranch := update.UpdatesByBranch{}
	byBranch.AddClosed(baseBranch, mockUpdate)

	assert.Equal(t, update.UpdatesByBranch{
		baseBranch: {
			Closed: []update.Update{mockUpdate},
		},
	}, byBranch)
}

func TestUpdates_OpenUpdate(t *testing.T) {
	u := update.Updates{Open: []update.Update{mockUpdate}}

	cases := map[string]struct {
		Update update.Update
		Open   string
	}{
		"open": {
			Update: mockUpdate,
			Open:   mockUpdate.Next,
		},
		"higher version": {
			Update: update.Update{Path: mockPath, Next: "v3.0.0"},
			Open:   mockUpdate.Next,
		},
		"lower version": {
			Update: update.Update{Path: mockPath, Next: "v1.0.0"},
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
	u := update.Updates{Open: []update.Update{mockUpdate}}
	assert.Equal(t, "", u.ClosedUpdate(mockUpdate))
}
