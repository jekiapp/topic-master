package group

import (
	"context"
	"errors"
	"testing"

	"github.com/jekiapp/topic-master/internal/model/acl"
)

type mockGroupListSimpleRepo struct {
	GetAllGroupsFunc func() ([]acl.Group, error)
}

func (m *mockGroupListSimpleRepo) GetAllGroups() ([]acl.Group, error) {
	return m.GetAllGroupsFunc()
}

func TestGetGroupListSimpleUsecase_Handle(t *testing.T) {
	tests := []struct {
		name     string
		mockRepo *mockGroupListSimpleRepo
		wantErr  bool
		wantLen  int
	}{
		{
			name: "repo error",
			mockRepo: &mockGroupListSimpleRepo{
				GetAllGroupsFunc: func() ([]acl.Group, error) { return nil, errors.New("fail") },
			},
			wantErr: true,
		},
		{
			name: "success",
			mockRepo: &mockGroupListSimpleRepo{
				GetAllGroupsFunc: func() ([]acl.Group, error) {
					return []acl.Group{{ID: "g1", Name: "n1"}}, nil
				},
			},
			wantLen: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := GetGroupListSimpleUsecase{groupRepo: tt.mockRepo}
			resp, err := uc.Handle(context.Background(), GetGroupListSimpleRequest{})
			if (err != nil) != tt.wantErr {
				t.Errorf("Handle() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(resp.Groups) != tt.wantLen {
				t.Errorf("expected %d groups, got %d", tt.wantLen, len(resp.Groups))
			}
		})
	}
}
