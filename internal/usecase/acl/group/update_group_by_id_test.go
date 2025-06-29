package group

import (
	"context"
	"testing"
	"time"

	"github.com/jekiapp/topic-master/internal/model/acl"
	group_mock "github.com/jekiapp/topic-master/internal/usecase/acl/group/mock"
	"github.com/jekiapp/topic-master/pkg/util"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestUpdateGroupByIDUsecase_Handle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	rootGroup := []acl.GroupRole{{GroupName: acl.GroupRoot}}
	nonRootGroup := []acl.GroupRole{{GroupName: "notroot"}}
	tests := []struct {
		name      string
		groups    []acl.GroupRole
		setupMock func(m *group_mock.MockiUpdateGroupRepo)
		req       UpdateGroupByIDRequest
		wantErr   bool
	}{
		{
			name:      "unauthorized user cannot update group",
			groups:    nil,
			setupMock: func(m *group_mock.MockiUpdateGroupRepo) {},
			req:       UpdateGroupByIDRequest{ID: "alice-group-id", Description: "Updated Description"},
			wantErr:   true,
		},
		{
			name:      "non-root user cannot update group",
			groups:    nonRootGroup,
			setupMock: func(m *group_mock.MockiUpdateGroupRepo) {},
			req:       UpdateGroupByIDRequest{ID: "bob-group-id", Description: "Updated Description"},
			wantErr:   true,
		},
		{
			name:      "missing group id should fail",
			groups:    rootGroup,
			setupMock: func(m *group_mock.MockiUpdateGroupRepo) {},
			req:       UpdateGroupByIDRequest{ID: "", Description: "Updated Description"},
			wantErr:   true,
		},
		{
			name:   "get group error for carol",
			groups: rootGroup,
			setupMock: func(m *group_mock.MockiUpdateGroupRepo) {
				m.EXPECT().GetGroupByID(
					"carol-group-id",
				).Return(
					acl.Group{},
					context.DeadlineExceeded,
				)
			},
			req:     UpdateGroupByIDRequest{ID: "carol-group-id", Description: "Updated Description"},
			wantErr: true,
		},
		{
			name:   "update error for dave",
			groups: rootGroup,
			setupMock: func(m *group_mock.MockiUpdateGroupRepo) {
				m.EXPECT().GetGroupByID(
					"dave-group-id",
				).Return(
					acl.Group{ID: "dave-group-id", Description: "Previous Description"},
					nil,
				)
				m.EXPECT().UpdateGroup(
					gomock.Any(),
				).Return(
					context.DeadlineExceeded,
				)
			},
			req:     UpdateGroupByIDRequest{ID: "dave-group-id", Description: "Updated Description"},
			wantErr: true,
		},
		{
			name:   "successfully update engineering group",
			groups: rootGroup,
			setupMock: func(m *group_mock.MockiUpdateGroupRepo) {
				m.EXPECT().GetGroupByID(
					"engineering-id",
				).Return(
					acl.Group{ID: "engineering-id", Description: "Previous Description", UpdatedAt: time.Now()},
					nil,
				)
				m.EXPECT().UpdateGroup(
					gomock.Any(),
				).Return(
					nil,
				)
			},
			req:     UpdateGroupByIDRequest{ID: "engineering-id", Description: "Updated Description"},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ctx context.Context = context.Background()
			if tt.groups != nil {
				user := &acl.User{
					ID:       "test-user",
					Username: "testuser",
					Groups:   tt.groups,
				}
				ctx = util.MockContextWithUser(ctx, user)
			}
			mockRepo := group_mock.NewMockiUpdateGroupRepo(ctrl)
			tt.setupMock(mockRepo)
			uc := UpdateGroupByIDUsecase{repo: mockRepo}
			_, err := uc.Handle(ctx, tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
