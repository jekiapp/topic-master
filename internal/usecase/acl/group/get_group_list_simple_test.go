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
			name: "repo returns error for alice's group",
			setupMock: func(m *group_mock.MockiGroupListSimpleRepo) {
				m.EXPECT().GetAllGroups().Return(
					nil,
					context.DeadlineExceeded,
				)
			},
			wantErr: true,
		},
		{
			name: "success for bob's group",
			setupMock: func(m *group_mock.MockiGroupListSimpleRepo) {
				m.EXPECT().GetAllGroups().Return(
					[]acl.Group{{ID: "bob-group", Name: "Bob Group"}},
					nil,
				)
			},
			wantLen: 1,
		},
		{
			name: "group with empty name for carol",
			setupMock: func(m *group_mock.MockiGroupListSimpleRepo) {
				m.EXPECT().GetAllGroups().Return(
					[]acl.Group{{ID: "carol-group", Name: ""}},
					nil,
				)
			},
			wantLen: 1,
		},
		{
			name: "repo returns empty slice for dave",
			setupMock: func(m *group_mock.MockiGroupListSimpleRepo) {
				m.EXPECT().GetAllGroups().Return(
					[]acl.Group{},
					nil,
				)
			},
			wantLen: 0,
		},
		{
			name: "repo returns error with non-empty slice for eve",
			setupMock: func(m *group_mock.MockiGroupListSimpleRepo) {
				m.EXPECT().GetAllGroups().Return(
					[]acl.Group{{ID: "eve-group", Name: "Eve Group"}},
					context.DeadlineExceeded,
				)
			},
			wantErr: true,
		},
		{
			name: "multiple groups (alice, bob)",
			setupMock: func(m *group_mock.MockiGroupListSimpleRepo) {
				m.EXPECT().GetAllGroups().Return(
					[]acl.Group{{ID: "alice-group", Name: "Alice Group"}, {ID: "bob-group", Name: "Bob Group"}},
					nil,
				)
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
