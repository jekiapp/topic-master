package group

import (
	"context"
	"testing"

	"github.com/jekiapp/topic-master/internal/model/acl"
	group_mock "github.com/jekiapp/topic-master/internal/usecase/acl/group/mock"
	"github.com/jekiapp/topic-master/pkg/util"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestDeleteGroupUsecase_Handle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	rootGroup := []acl.GroupRole{{GroupName: acl.GroupRoot}}
	nonRootGroup := []acl.GroupRole{{GroupName: "notroot"}}
	tests := []struct {
		name      string
		groups    []acl.GroupRole
		setupMock func(m *group_mock.MockiDeleteGroupRepo)
		req       DeleteGroupRequest
		wantErr   bool
		wantOK    bool
	}{
		{
			name:      "unauthorized user cannot delete group",
			groups:    nil,
			setupMock: func(m *group_mock.MockiDeleteGroupRepo) {},
			req:       DeleteGroupRequest{ID: "alice-group-id"},
			wantErr:   true,
		},
		{
			name:      "non-root user cannot delete group",
			groups:    nonRootGroup,
			setupMock: func(m *group_mock.MockiDeleteGroupRepo) {},
			req:       DeleteGroupRequest{ID: "bob-group-id"},
			wantErr:   true,
		},
		{
			name:      "missing group id should fail",
			groups:    rootGroup,
			setupMock: func(m *group_mock.MockiDeleteGroupRepo) {},
			req:       DeleteGroupRequest{},
			wantErr:   true,
		},
		{
			name:   "group not found for carol",
			groups: rootGroup,
			setupMock: func(m *group_mock.MockiDeleteGroupRepo) {
				m.EXPECT().GetGroupByID(
					"carol-group-id",
				).Return(
					acl.Group{},
					context.DeadlineExceeded,
				)
			},
			req:     DeleteGroupRequest{ID: "carol-group-id"},
			wantErr: true,
		},
		{
			name:   "delete root group should fail",
			groups: rootGroup,
			setupMock: func(m *group_mock.MockiDeleteGroupRepo) {
				m.EXPECT().GetGroupByID(
					"root-group-id",
				).Return(
					acl.Group{Name: acl.GroupRoot},
					nil,
				)
			},
			req:     DeleteGroupRequest{ID: "root-group-id"},
			wantErr: true,
		},
		{
			name:   "repo delete error for engineering",
			groups: rootGroup,
			setupMock: func(m *group_mock.MockiDeleteGroupRepo) {
				m.EXPECT().GetGroupByID(
					"engineering-id",
				).Return(
					acl.Group{Name: "notroot"},
					nil,
				)
				m.EXPECT().DeleteGroupByID(
					"engineering-id",
				).Return(
					context.DeadlineExceeded,
				)
			},
			req:     DeleteGroupRequest{ID: "engineering-id"},
			wantErr: true,
		},
		{
			name:   "successfully delete marketing group",
			groups: rootGroup,
			setupMock: func(m *group_mock.MockiDeleteGroupRepo) {
				m.EXPECT().GetGroupByID(
					"marketing-id",
				).Return(
					acl.Group{Name: "notroot"},
					nil,
				)
				m.EXPECT().DeleteGroupByID(
					"marketing-id",
				).Return(
					nil,
				)
			},
			req:    DeleteGroupRequest{ID: "marketing-id"},
			wantOK: true,
		},
		{
			name:   "GetGroupByID returns group with empty name for support",
			groups: rootGroup,
			setupMock: func(m *group_mock.MockiDeleteGroupRepo) {
				m.EXPECT().GetGroupByID(
					"support-id",
				).Return(
					acl.Group{},
					nil,
				)
				m.EXPECT().DeleteGroupByID(
					"support-id",
				).Return(
					nil,
				)
			},
			req:    DeleteGroupRequest{ID: "support-id"},
			wantOK: true,
		},
		{
			name:   "DeleteGroupByID returns nil but GetGroupByID returns error for devops",
			groups: rootGroup,
			setupMock: func(m *group_mock.MockiDeleteGroupRepo) {
				m.EXPECT().GetGroupByID(
					"devops-id",
				).Return(
					acl.Group{},
					context.DeadlineExceeded,
				)
				m.EXPECT().DeleteGroupByID(
					"devops-id",
				).Return(
					nil,
				)
			},
			req:     DeleteGroupRequest{ID: "devops-id"},
			wantErr: true,
		},
		{
			name:   "forbidden group name edge case for finance",
			groups: rootGroup,
			setupMock: func(m *group_mock.MockiDeleteGroupRepo) {
				m.EXPECT().GetGroupByID(
					"finance-id",
				).Return(
					acl.Group{Name: "forbidden"},
					nil,
				)
				m.EXPECT().DeleteGroupByID(
					"finance-id",
				).Return(
					context.DeadlineExceeded,
				)
			},
			req:     DeleteGroupRequest{ID: "finance-id"},
			wantErr: true,
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
			mockRepo := group_mock.NewMockiDeleteGroupRepo(ctrl)
			tt.setupMock(mockRepo)
			uc := DeleteGroupUsecase{repo: mockRepo}
			resp, err := uc.Handle(ctx, tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if tt.wantOK {
				assert.True(t, resp.Success)
			}
		})
	}
}
