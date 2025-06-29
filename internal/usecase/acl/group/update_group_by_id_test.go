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
			name:      "unauthorized",
			groups:    nil,
			setupMock: func(m *group_mock.MockiUpdateGroupRepo) {},
			req:       UpdateGroupByIDRequest{ID: "dev-group-id", Description: "updated-desc"},
			wantErr:   true,
		},
		{
			name:      "not root",
			groups:    nonRootGroup,
			setupMock: func(m *group_mock.MockiUpdateGroupRepo) {},
			req:       UpdateGroupByIDRequest{ID: "dev-group-id", Description: "updated-desc"},
			wantErr:   true,
		},
		{
			name:      "missing id",
			groups:    rootGroup,
			setupMock: func(m *group_mock.MockiUpdateGroupRepo) {},
			req:       UpdateGroupByIDRequest{ID: "", Description: "updated-desc"},
			wantErr:   true,
		},
		{
			name:   "get group error",
			groups: rootGroup,
			setupMock: func(m *group_mock.MockiUpdateGroupRepo) {
				m.EXPECT().GetGroupByID("qa-team-id").Return(acl.Group{}, context.DeadlineExceeded)
			},
			req:     UpdateGroupByIDRequest{ID: "qa-team-id", Description: "updated-desc"},
			wantErr: true,
		},
		{
			name:   "update error",
			groups: rootGroup,
			setupMock: func(m *group_mock.MockiUpdateGroupRepo) {
				m.EXPECT().GetGroupByID("hr-team-id").Return(acl.Group{ID: "hr-team-id", Description: "previous-desc"}, nil)
				m.EXPECT().UpdateGroup(gomock.Any()).Return(context.DeadlineExceeded)
			},
			req:     UpdateGroupByIDRequest{ID: "hr-team-id", Description: "updated-desc"},
			wantErr: true,
		},
		{
			name:   "success",
			groups: rootGroup,
			setupMock: func(m *group_mock.MockiUpdateGroupRepo) {
				m.EXPECT().GetGroupByID("ops-team-id").Return(acl.Group{ID: "ops-team-id", Description: "previous-desc", UpdatedAt: time.Now()}, nil)
				m.EXPECT().UpdateGroup(gomock.Any()).Return(nil)
			},
			req:     UpdateGroupByIDRequest{ID: "ops-team-id", Description: "updated-desc"},
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
