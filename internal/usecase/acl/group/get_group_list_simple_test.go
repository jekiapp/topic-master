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
		{
			name: "group with empty name",
			mockRepo: &mockGroupListSimpleRepo{
				GetAllGroupsFunc: func() ([]acl.Group, error) {
					return []acl.Group{{ID: "g2", Name: ""}}, nil
				},
			},
			wantLen: 1,
		},
		{
			name: "repo returns empty slice",
			mockRepo: &mockGroupListSimpleRepo{
				GetAllGroupsFunc: func() ([]acl.Group, error) { return []acl.Group{}, nil },
			},
			wantLen: 0,
		},
		{
			name: "repo returns error with non-empty slice",
			mockRepo: &mockGroupListSimpleRepo{
				GetAllGroupsFunc: func() ([]acl.Group, error) { return []acl.Group{{ID: "g3", Name: "n3"}}, errors.New("fail") },
			},
			wantErr: true,
		},
		{
			name: "multiple groups",
			mockRepo: &mockGroupListSimpleRepo{
				GetAllGroupsFunc: func() ([]acl.Group, error) {
					return []acl.Group{
						{ID: "g1", Name: "n1"},
						{ID: "g2", Name: "n2"},
					}, nil
				},
			},
			wantLen: 2,
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
