package group

import (
	"context"
	"testing"

	"github.com/jekiapp/topic-master/internal/model/acl"
	group_mock "github.com/jekiapp/topic-master/internal/usecase/acl/group/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

type ctxKeyUserInfo struct{}

func jwtClaimsForGroups(groups []acl.GroupRole) *acl.JWTClaims {
	return &acl.JWTClaims{
		UserID:   "test-user",
		Username: "testuser",
		Groups:   groups,
	}
}

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
			name:      "unauthorized",
			groups:    nil,
			setupMock: func(m *group_mock.MockiDeleteGroupRepo) {},
			req:       DeleteGroupRequest{ID: "dev-group-id"},
			wantErr:   true,
		},
		{
			name:      "not root",
			groups:    nonRootGroup,
			setupMock: func(m *group_mock.MockiDeleteGroupRepo) {},
			req:       DeleteGroupRequest{ID: "dev-group-id"},
			wantErr:   true,
		},
		{
			name:      "missing id",
			groups:    rootGroup,
			setupMock: func(m *group_mock.MockiDeleteGroupRepo) {},
			req:       DeleteGroupRequest{},
			wantErr:   true,
		},
		{
			name:   "group not found",
			groups: rootGroup,
			setupMock: func(m *group_mock.MockiDeleteGroupRepo) {
				m.EXPECT().GetGroupByID("qa-team-id").Return(acl.Group{}, context.DeadlineExceeded)
			},
			req:     DeleteGroupRequest{ID: "qa-team-id"},
			wantErr: true,
		},
		{
			name:   "delete root group",
			groups: rootGroup,
			setupMock: func(m *group_mock.MockiDeleteGroupRepo) {
				m.EXPECT().GetGroupByID("hr-team-id").Return(acl.Group{Name: acl.GroupRoot}, nil)
			},
			req:     DeleteGroupRequest{ID: "hr-team-id"},
			wantErr: true,
		},
		{
			name:   "repo delete error",
			groups: rootGroup,
			setupMock: func(m *group_mock.MockiDeleteGroupRepo) {
				m.EXPECT().GetGroupByID("ops-team-id").Return(acl.Group{Name: "notroot"}, nil)
				m.EXPECT().DeleteGroupByID("ops-team-id").Return(context.DeadlineExceeded)
			},
			req:     DeleteGroupRequest{ID: "ops-team-id"},
			wantErr: true,
		},
		{
			name:   "success",
			groups: rootGroup,
			setupMock: func(m *group_mock.MockiDeleteGroupRepo) {
				m.EXPECT().GetGroupByID("eng-team-id").Return(acl.Group{Name: "notroot"}, nil)
				m.EXPECT().DeleteGroupByID("eng-team-id").Return(nil)
			},
			req:    DeleteGroupRequest{ID: "eng-team-id"},
			wantOK: true,
		},
		{
			name:   "GetGroupByID returns group with empty name",
			groups: rootGroup,
			setupMock: func(m *group_mock.MockiDeleteGroupRepo) {
				m.EXPECT().GetGroupByID("marketing-id").Return(acl.Group{}, nil)
				m.EXPECT().DeleteGroupByID("marketing-id").Return(nil)
			},
			req:    DeleteGroupRequest{ID: "marketing-id"},
			wantOK: true,
		},
		{
			name:   "DeleteGroupByID returns nil but GetGroupByID returns error",
			groups: rootGroup,
			setupMock: func(m *group_mock.MockiDeleteGroupRepo) {
				m.EXPECT().GetGroupByID("support-id").Return(acl.Group{}, context.DeadlineExceeded)
				m.EXPECT().DeleteGroupByID("support-id").Return(nil)
			},
			req:     DeleteGroupRequest{ID: "support-id"},
			wantErr: true,
		},
		{
			name:   "forbidden group name edge case",
			groups: rootGroup,
			setupMock: func(m *group_mock.MockiDeleteGroupRepo) {
				m.EXPECT().GetGroupByID("forbidden-id").Return(acl.Group{Name: "forbidden"}, nil)
				m.EXPECT().DeleteGroupByID("forbidden-id").Return(context.DeadlineExceeded)
			},
			req:     DeleteGroupRequest{ID: "forbidden-id"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.groups != nil {
				ctx = context.WithValue(ctx, ctxKeyUserInfo{}, jwtClaimsForGroups(tt.groups))
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
