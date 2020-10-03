package updater_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	updater2 "github.com/thepwagner/action-update/updater"
)

//go:generate mockery --outpkg updater_test --output . --testonly --name Updater --structname mockUpdater --filename mockupdater_test.go
//go:generate mockery --outpkg updater_test --output . --testonly --name Repo --structname mockRepo --filename mockrepo_test.go

func TestRepoUpdater_Update(t *testing.T) {
	r := &mockRepo{}
	u := &mockUpdater{}
	ru := updater2.NewRepoUpdater(r, u)
	ctx := context.Background()

	branch := setupMockUpdate(ctx, r, u, mockUpdate)

	err := ru.Update(ctx, baseBranch, branch, mockUpdate)
	require.NoError(t, err)
}

func setupMockUpdate(ctx context.Context, r *mockRepo, u *mockUpdater, up updater2.Update) string {
	branch := fmt.Sprintf("action-update-go/main/%s/%s", up.Path, up.Next)
	r.On("NewBranch", baseBranch, branch).Return(nil)
	u.On("ApplyUpdate", ctx, up).Return(nil)
	r.On("Push", ctx, up).Return(nil)
	return branch
}

func TestRepoUpdater_UpdateAll_NoChanges(t *testing.T) {
	r := &mockRepo{}
	u := &mockUpdater{}
	ru := updater2.NewRepoUpdater(r, u)
	ctx := context.Background()

	r.On("SetBranch", baseBranch).Return(nil)
	dep := updater2.Dependency{Path: mockUpdate.Path, Version: mockUpdate.Previous}
	u.On("Dependencies", ctx).Return([]updater2.Dependency{dep}, nil)
	u.On("Check", ctx, dep).Return(nil, nil)

	err := ru.UpdateAll(ctx, baseBranch)
	require.NoError(t, err)
	r.AssertExpectations(t)
	u.AssertExpectations(t)
}

func TestRepoUpdater_UpdateAll_Update(t *testing.T) {
	r := &mockRepo{}
	u := &mockUpdater{}
	ru := updater2.NewRepoUpdater(r, u)
	ctx := context.Background()

	r.On("SetBranch", baseBranch).Return(nil)
	dep := updater2.Dependency{Path: mockUpdate.Path, Version: mockUpdate.Previous}
	u.On("Dependencies", ctx).Return([]updater2.Dependency{dep}, nil)
	availableUpdate := mockUpdate // avoid pointer to shared reference
	u.On("Check", ctx, dep).Return(&availableUpdate, nil)
	setupMockUpdate(ctx, r, u, mockUpdate) // delegates to .Update()

	err := ru.UpdateAll(ctx, baseBranch)
	require.NoError(t, err)
	r.AssertExpectations(t)
	u.AssertExpectations(t)
}

func TestRepoUpdater_UpdateAll_Multiple(t *testing.T) {
	r := &mockRepo{}
	u := &mockUpdater{}
	ru := updater2.NewRepoUpdater(r, u)
	ctx := context.Background()

	r.On("SetBranch", baseBranch).Return(nil)
	dep := updater2.Dependency{Path: mockUpdate.Path, Version: mockUpdate.Previous}
	otherDep := updater2.Dependency{Path: "github.com/foo/baz", Version: mockUpdate.Previous}
	u.On("Dependencies", ctx).Return([]updater2.Dependency{dep, otherDep}, nil)
	availableUpdate := mockUpdate // avoid pointer to shared reference
	u.On("Check", ctx, dep).Return(&availableUpdate, nil)
	otherUpdate := updater2.Update{Path: "github.com/foo/baz", Next: "v3.0.0"}
	u.On("Check", ctx, otherDep).Return(&otherUpdate, nil)
	setupMockUpdate(ctx, r, u, mockUpdate)  // delegates to .Update()
	setupMockUpdate(ctx, r, u, otherUpdate) // delegates to .Update()

	err := ru.UpdateAll(ctx, baseBranch)
	require.NoError(t, err)
	r.AssertExpectations(t)
	u.AssertExpectations(t)
}

func TestRepoUpdater_UpdateAll_MultipleBatch(t *testing.T) {
	r := &mockRepo{}
	u := &mockUpdater{}
	batchName := "foo"
	ru := updater2.NewRepoUpdater(r, u, updater2.WithBatches(map[string][]string{batchName: {"github.com/foo"}}))
	ctx := context.Background()

	r.On("SetBranch", baseBranch).Return(nil)
	dep := updater2.Dependency{Path: mockUpdate.Path, Version: mockUpdate.Previous}
	otherDep := updater2.Dependency{Path: "github.com/foo/baz", Version: mockUpdate.Previous}
	u.On("Dependencies", ctx).Return([]updater2.Dependency{dep, otherDep}, nil)
	availableUpdate := mockUpdate // avoid pointer to shared reference
	u.On("Check", ctx, dep).Return(&availableUpdate, nil)
	otherUpdate := updater2.Update{Path: "github.com/foo/baz", Next: "v3.0.0"}
	u.On("Check", ctx, otherDep).Return(&otherUpdate, nil)

	r.On("NewBranch", baseBranch, "action-update-go/main/foo").Times(1).Return(nil)
	u.On("ApplyUpdate", ctx, mock.Anything).Times(2).Return(nil)
	r.On("Push", ctx, mock.Anything, mock.Anything).Times(1).Return(nil)

	err := ru.UpdateAll(ctx, baseBranch)
	require.NoError(t, err)
	r.AssertExpectations(t)
	u.AssertExpectations(t)
}
