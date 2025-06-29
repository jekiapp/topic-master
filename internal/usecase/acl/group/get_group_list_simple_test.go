package group

import (
	"context"
	"testing"

	"github.com/jekiapp/topic-master/internal/model/acl"
	group_mock "github.com/jekiapp/topic-master/internal/usecase/acl/group/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestGetGroupListSimpleUsecase_Handle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name      string
		setupMock func(m *group_mock.MockiGroupListSimpleRepo)
		wantErr   bool
		wantLen   int
	}{
		{
			name: "repo error",
			setupMock: func(m *group_mock.MockiGroupListSimpleRepo) {
				m.EXPECT().GetAllGroups().Return(nil, context.DeadlineExceeded)
			},
			wantErr: true,
		},
		{
			name: "success",
			setupMock: func(m *group_mock.MockiGroupListSimpleRepo) {
				m.EXPECT().GetAllGroups().Return([]acl.Group{{ID: "dev-group", Name: "Dev Group"}}, nil)
			},
			wantLen: 1,
		},
		{
			name: "group with empty name",
			setupMock: func(m *group_mock.MockiGroupListSimpleRepo) {
				m.EXPECT().GetAllGroups().Return([]acl.Group{{ID: "qa-team", Name: ""}}, nil)
			},
			wantLen: 1,
		},
		{
			name: "repo returns empty slice",
			setupMock: func(m *group_mock.MockiGroupListSimpleRepo) {
				m.EXPECT().GetAllGroups().Return([]acl.Group{}, nil)
			},
			wantLen: 0,
		},
		{
			name: "repo returns error with non-empty slice",
			setupMock: func(m *group_mock.MockiGroupListSimpleRepo) {
				m.EXPECT().GetAllGroups().Return([]acl.Group{{ID: "ops-team", Name: "Ops Team"}}, context.DeadlineExceeded)
			},
			wantErr: true,
		},
		{
			name: "multiple groups",
			setupMock: func(m *group_mock.MockiGroupListSimpleRepo) {
				m.EXPECT().GetAllGroups().Return([]acl.Group{{ID: "dev-group", Name: "Dev Group"}, {ID: "qa-team", Name: "QA Team"}}, nil)
			},
			wantLen: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := group_mock.NewMockiGroupListSimpleRepo(ctrl)
			tt.setupMock(mockRepo)
			uc := GetGroupListSimpleUsecase{groupRepo: mockRepo}
			resp, err := uc.Handle(context.Background(), GetGroupListSimpleRequest{})
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantLen, len(resp.Groups))
		})
	}
}
