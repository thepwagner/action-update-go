package update_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/thepwagner/action-update-go/update"
)

//go:generate mockery --outpkg update_test --output . --testonly --name Updater --structname mockUpdater --filename mockupdater_test.go
//go:generate mockery --outpkg update_test --output . --testonly --name Repo --structname mockRepo --filename mockrepo_test.go

func TestRepoUpdater_Update(t *testing.T) {
	r := &mockRepo{}
	u := &mockUpdater{}
	ru := update.NewRepoUpdater(r, u)
	ctx := context.Background()

	setupMockUpdate(ctx, r, u, mockUpdate)

	err := ru.Update(ctx, baseBranch, mockUpdate)
	require.NoError(t, err)
}

func setupMockUpdate(ctx context.Context, r *mockRepo, u *mockUpdater, up update.Update) {
	r.On("NewBranch", baseBranch, up).Return(nil)
	u.On("ApplyUpdate", ctx, up).Return(nil)
	r.On("Push", ctx, up).Return(nil)
}

func TestRepoUpdater_UpdateAll_NoChanges(t *testing.T) {
	r := &mockRepo{}
	u := &mockUpdater{}
	ru := update.NewRepoUpdater(r, u)
	ctx := context.Background()

	r.On("Updates", ctx).Return(update.UpdatesByBranch{}, nil)
	r.On("SetBranch", baseBranch).Return(nil)
	dep := update.Dependency{Path: mockUpdate.Path, Version: mockUpdate.Previous}
	u.On("Dependencies", ctx).Return([]update.Dependency{dep}, nil)
	u.On("Check", ctx, dep).Return(nil, nil)

	err := ru.UpdateAll(ctx, baseBranch)
	require.NoError(t, err)
	r.AssertExpectations(t)
	u.AssertExpectations(t)
}

func TestRepoUpdater_UpdateAll_Update(t *testing.T) {
	r := &mockRepo{}
	u := &mockUpdater{}
	ru := update.NewRepoUpdater(r, u)
	ctx := context.Background()

	r.On("Updates", ctx).Return(update.UpdatesByBranch{}, nil)
	r.On("SetBranch", baseBranch).Return(nil)
	dep := update.Dependency{Path: mockUpdate.Path, Version: mockUpdate.Previous}
	u.On("Dependencies", ctx).Return([]update.Dependency{dep}, nil)
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
	ru := update.NewRepoUpdater(r, u)
	ctx := context.Background()

	r.On("Updates", ctx).Return(update.UpdatesByBranch{}, nil)
	r.On("SetBranch", baseBranch).Return(nil)
	dep := update.Dependency{Path: mockUpdate.Path, Version: mockUpdate.Previous}
	otherDep := update.Dependency{Path: "github.com/foo/baz", Version: mockUpdate.Previous}
	u.On("Dependencies", ctx).Return([]update.Dependency{dep, otherDep}, nil)
	availableUpdate := mockUpdate // avoid pointer to shared reference
	u.On("Check", ctx, dep).Return(&availableUpdate, nil)
	otherUpdate := update.Update{Path: "github.com/foo/baz", Next: "v3.0.0"}
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
	ru := update.NewRepoUpdater(r, u)
	ru.Batch = true
	ctx := context.Background()

	r.On("Updates", ctx).Return(update.UpdatesByBranch{}, nil)
	r.On("SetBranch", baseBranch).Return(nil)
	dep := update.Dependency{Path: mockUpdate.Path, Version: mockUpdate.Previous}
	otherDep := update.Dependency{Path: "github.com/foo/baz", Version: mockUpdate.Previous}
	u.On("Dependencies", ctx).Return([]update.Dependency{dep, otherDep}, nil)
	availableUpdate := mockUpdate // avoid pointer to shared reference
	u.On("Check", ctx, dep).Return(&availableUpdate, nil)
	otherUpdate := update.Update{Path: "github.com/foo/baz", Next: "v3.0.0"}
	u.On("Check", ctx, otherDep).Return(&otherUpdate, nil)

	r.On("NewBranch", baseBranch, mock.Anything).Times(1).Return(nil)
	u.On("ApplyUpdate", ctx, mock.Anything).Times(2).Return(nil)
	r.On("Push", ctx, mock.Anything).Times(1).Return(nil)

	err := ru.UpdateAll(ctx, baseBranch)
	require.NoError(t, err)
	r.AssertExpectations(t)
	u.AssertExpectations(t)
}
