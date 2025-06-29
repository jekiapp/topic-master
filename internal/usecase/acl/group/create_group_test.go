package group

import (
	"context"
	"testing"

	"github.com/jekiapp/topic-master/internal/model/acl"
	group_mock "github.com/jekiapp/topic-master/internal/usecase/acl/group/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestCreateGroupUsecase_Handle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name      string
		req       CreateGroupRequest
		setupMock func(m *group_mock.MockiCreateGroupRepo)
		wantErr   bool
		wantGroup bool
	}{
		{
			name:      "missing group name should fail",
			req:       CreateGroupRequest{},
			setupMock: func(m *group_mock.MockiCreateGroupRepo) {},
			wantErr:   true,
		},
		{
			name: "group already exists for alice",
			req:  CreateGroupRequest{Name: "alice-group"},
			setupMock: func(m *group_mock.MockiCreateGroupRepo) {
				m.EXPECT().GetGroupByName("alice-group").Return(acl.Group{}, nil)
			},
			wantErr: true,
		},
		{
			name: "repo create error for bob",
			req:  CreateGroupRequest{Name: "bob-team"},
			setupMock: func(m *group_mock.MockiCreateGroupRepo) {
				m.EXPECT().GetGroupByName("bob-team").Return(acl.Group{}, context.DeadlineExceeded)
				m.EXPECT().CreateGroup(gomock.Any()).Return(context.DeadlineExceeded)
			},
			wantErr: true,
		},
		{
			name: "success for engineering group",
			req:  CreateGroupRequest{Name: "engineering", Description: "Engineering Department"},
			setupMock: func(m *group_mock.MockiCreateGroupRepo) {
				m.EXPECT().GetGroupByName("engineering").Return(acl.Group{}, context.DeadlineExceeded)
				m.EXPECT().CreateGroup(gomock.Any()).Return(nil)
			},
			wantErr:   false,
			wantGroup: true,
		},
		{
			name: "special characters in group name",
			req:  CreateGroupRequest{Name: "!@#$$%", Description: "Special Group"},
			setupMock: func(m *group_mock.MockiCreateGroupRepo) {
				m.EXPECT().GetGroupByName("!@#$$%").Return(acl.Group{}, context.DeadlineExceeded)
				m.EXPECT().CreateGroup(gomock.Any()).Return(nil)
			},
			wantErr:   false,
			wantGroup: true,
		},
		{
			name: "repo GetGroupByName returns error other than not found for carol",
			req:  CreateGroupRequest{Name: "carol-team"},
			setupMock: func(m *group_mock.MockiCreateGroupRepo) {
				m.EXPECT().GetGroupByName("carol-team").Return(acl.Group{}, context.Canceled)
			},
			wantErr: true,
		},
		{
			name: "duplicate group for alice with different description",
			req:  CreateGroupRequest{Name: "alice-group", Description: "Updated Description"},
			setupMock: func(m *group_mock.MockiCreateGroupRepo) {
				m.EXPECT().GetGroupByName("alice-group").Return(acl.Group{Name: "alice-group", Description: "Previous Description"}, nil)
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := group_mock.NewMockiCreateGroupRepo(ctrl)
			tt.setupMock(mockRepo)
			uc := CreateGroupUsecase{repo: mockRepo}
			resp, err := uc.Handle(context.Background(), tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			if tt.wantGroup {
				assert.NotEmpty(t, resp.Group)
				assert.NotEmpty(t, resp.Group.CreatedAt)
				assert.NotEmpty(t, resp.Group.UpdatedAt)
			}
		})
	}
}
