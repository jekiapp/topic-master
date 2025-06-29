package group

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jekiapp/topic-master/internal/model"
	"github.com/jekiapp/topic-master/internal/model/acl"
)

type mockUpdateGroupRepo struct {
	GetGroupByIDFunc func(id string) (acl.Group, error)
	UpdateGroupFunc  func(group acl.Group) error
}

func (m *mockUpdateGroupRepo) GetGroupByID(id string) (acl.Group, error) {
	return m.GetGroupByIDFunc(id)
}
func (m *mockUpdateGroupRepo) UpdateGroup(group acl.Group) error {
	return m.UpdateGroupFunc(group)
}

func jwtClaimsForGroupsUpdate(groups []acl.GroupRole) *acl.JWTClaims {
	return &acl.JWTClaims{
		UserID:   "test-user",
		Username: "testuser",
		Groups:   groups,
	}
}

func TestUpdateGroupByIDUsecase_Handle(t *testing.T) {
	rootGroup := []acl.GroupRole{{GroupName: acl.GroupRoot}}
	nonRootGroup := []acl.GroupRole{{GroupName: "notroot"}}
	tests := []struct {
		name     string
		groups   []acl.GroupRole
		req      UpdateGroupByIDRequest
		mockRepo *mockUpdateGroupRepo
		wantErr  bool
	}{
		{
			name:     "unauthorized",
			groups:   nil,
			req:      UpdateGroupByIDRequest{ID: "id1", Description: "desc"},
			mockRepo: &mockUpdateGroupRepo{},
			wantErr:  true,
		},
		{
			name:     "not root",
			groups:   nonRootGroup,
			req:      UpdateGroupByIDRequest{ID: "id1", Description: "desc"},
			mockRepo: &mockUpdateGroupRepo{},
			wantErr:  true,
		},
		{
			name:     "missing id",
			groups:   rootGroup,
			req:      UpdateGroupByIDRequest{ID: "", Description: "desc"},
			mockRepo: &mockUpdateGroupRepo{},
			wantErr:  true,
		},
		{
			name:   "get group error",
			groups: rootGroup,
			req:    UpdateGroupByIDRequest{ID: "id2", Description: "desc"},
			mockRepo: &mockUpdateGroupRepo{
				GetGroupByIDFunc: func(id string) (acl.Group, error) { return acl.Group{}, errors.New("fail") },
			},
			wantErr: true,
		},
		{
			name:   "update error",
			groups: rootGroup,
			req:    UpdateGroupByIDRequest{ID: "id3", Description: "desc"},
			mockRepo: &mockUpdateGroupRepo{
				GetGroupByIDFunc: func(id string) (acl.Group, error) { return acl.Group{ID: "id3", Description: "old"}, nil },
				UpdateGroupFunc:  func(group acl.Group) error { return errors.New("fail") },
			},
			wantErr: true,
		},
		{
			name:   "success",
			groups: rootGroup,
			req:    UpdateGroupByIDRequest{ID: "id4", Description: "desc"},
			mockRepo: &mockUpdateGroupRepo{
				GetGroupByIDFunc: func(id string) (acl.Group, error) {
					return acl.Group{ID: "id4", Description: "old", UpdatedAt: time.Now()}, nil
				},
				UpdateGroupFunc: func(group acl.Group) error { return nil },
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.groups != nil {
				ctx = context.WithValue(ctx, model.UserInfoKey, jwtClaimsForGroupsUpdate(tt.groups))
			}
			uc := UpdateGroupByIDUsecase{repo: tt.mockRepo}
			_, err := uc.Handle(ctx, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handle() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
