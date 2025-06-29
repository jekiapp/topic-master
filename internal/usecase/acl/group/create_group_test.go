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
			name:      "missing name",
			req:       CreateGroupRequest{},
			setupMock: func(m *group_mock.MockiCreateGroupRepo) {},
			wantErr:   true,
		},
		{
			name: "group already exists",
			req:  CreateGroupRequest{Name: "dev-group"},
			setupMock: func(m *group_mock.MockiCreateGroupRepo) {
				m.EXPECT().GetGroupByName("dev-group").Return(acl.Group{}, nil)
			},
			wantErr: true,
		},
		{
			name: "repo create error",
			req:  CreateGroupRequest{Name: "qa-team"},
			setupMock: func(m *group_mock.MockiCreateGroupRepo) {
				m.EXPECT().GetGroupByName("qa-team").Return(acl.Group{}, context.DeadlineExceeded)
				m.EXPECT().CreateGroup(gomock.Any()).Return(context.DeadlineExceeded)
			},
			wantErr: true,
		},
		{
			name: "success",
			req:  CreateGroupRequest{Name: "hr-team", Description: "engineering"},
			setupMock: func(m *group_mock.MockiCreateGroupRepo) {
				m.EXPECT().GetGroupByName("hr-team").Return(acl.Group{}, context.DeadlineExceeded)
				m.EXPECT().CreateGroup(gomock.Any()).Return(nil)
			},
			wantErr:   false,
			wantGroup: true,
		},
		{
			name: "special characters in name",
			req:  CreateGroupRequest{Name: "!@#$$%", Description: "engineering"},
			setupMock: func(m *group_mock.MockiCreateGroupRepo) {
				m.EXPECT().GetGroupByName("!@#$$%").Return(acl.Group{}, context.DeadlineExceeded)
				m.EXPECT().CreateGroup(gomock.Any()).Return(nil)
			},
			wantErr:   false,
			wantGroup: true,
		},
		{
			name: "repo GetGroupByName returns error other than not found",
			req:  CreateGroupRequest{Name: "ops-team"},
			setupMock: func(m *group_mock.MockiCreateGroupRepo) {
				m.EXPECT().GetGroupByName("ops-team").Return(acl.Group{}, context.Canceled)
			},
			wantErr: true,
		},
		{
			name: "duplicate group with different description",
			req:  CreateGroupRequest{Name: "dev-group", Description: "updated-desc"},
			setupMock: func(m *group_mock.MockiCreateGroupRepo) {
				m.EXPECT().GetGroupByName("dev-group").Return(acl.Group{Name: "dev-group", Description: "previous-desc"}, nil)
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
