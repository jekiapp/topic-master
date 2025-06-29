package group

import (
	"context"
	"errors"
	"testing"

	"github.com/jekiapp/topic-master/internal/model"
	"github.com/jekiapp/topic-master/internal/model/acl"
)

type mockDeleteGroupRepo struct {
	DeleteGroupByIDFunc func(id string) error
	GetGroupByIDFunc    func(id string) (acl.Group, error)
}

func (m *mockDeleteGroupRepo) DeleteGroupByID(id string) error {
	return m.DeleteGroupByIDFunc(id)
}
func (m *mockDeleteGroupRepo) GetGroupByID(id string) (acl.Group, error) {
	return m.GetGroupByIDFunc(id)
}

func jwtClaimsForGroups(groups []acl.GroupRole) *acl.JWTClaims {
	return &acl.JWTClaims{
		UserID:   "test-user",
		Username: "testuser",
		Groups:   groups,
	}
}

func TestDeleteGroupUsecase_Handle(t *testing.T) {
	rootGroup := []acl.GroupRole{{GroupName: acl.GroupRoot}}
	nonRootGroup := []acl.GroupRole{{GroupName: "notroot"}}
	tests := []struct {
		name     string
		groups   []acl.GroupRole
		req      DeleteGroupRequest
		mockRepo *mockDeleteGroupRepo
		wantErr  bool
		wantOK   bool
	}{
		{
			name:     "unauthorized",
			groups:   nil,
			req:      DeleteGroupRequest{ID: "id1"},
			mockRepo: &mockDeleteGroupRepo{},
			wantErr:  true,
		},
		{
			name:     "not root",
			groups:   nonRootGroup,
			req:      DeleteGroupRequest{ID: "id1"},
			mockRepo: &mockDeleteGroupRepo{},
			wantErr:  true,
		},
		{
			name:     "missing id",
			groups:   rootGroup,
			req:      DeleteGroupRequest{},
			mockRepo: &mockDeleteGroupRepo{},
			wantErr:  true,
		},
		{
			name:   "group not found",
			groups: rootGroup,
			req:    DeleteGroupRequest{ID: "id2"},
			mockRepo: &mockDeleteGroupRepo{
				GetGroupByIDFunc: func(id string) (acl.Group, error) { return acl.Group{}, errors.New("not found") },
			},
			wantErr: true,
		},
		{
			name:   "delete root group",
			groups: rootGroup,
			req:    DeleteGroupRequest{ID: "id3"},
			mockRepo: &mockDeleteGroupRepo{
				GetGroupByIDFunc: func(id string) (acl.Group, error) { return acl.Group{Name: acl.GroupRoot}, nil },
			},
			wantErr: true,
		},
		{
			name:   "repo delete error",
			groups: rootGroup,
			req:    DeleteGroupRequest{ID: "id4"},
			mockRepo: &mockDeleteGroupRepo{
				GetGroupByIDFunc:    func(id string) (acl.Group, error) { return acl.Group{Name: "notroot"}, nil },
				DeleteGroupByIDFunc: func(id string) error { return errors.New("fail") },
			},
			wantErr: true,
		},
		{
			name:   "success",
			groups: rootGroup,
			req:    DeleteGroupRequest{ID: "id5"},
			mockRepo: &mockDeleteGroupRepo{
				GetGroupByIDFunc:    func(id string) (acl.Group, error) { return acl.Group{Name: "notroot"}, nil },
				DeleteGroupByIDFunc: func(id string) error { return nil },
			},
			wantOK: true,
		},
		{
			name:   "GetGroupByID returns group with empty name",
			groups: rootGroup,
			req:    DeleteGroupRequest{ID: "id6"},
			mockRepo: &mockDeleteGroupRepo{
				GetGroupByIDFunc:    func(id string) (acl.Group, error) { return acl.Group{}, nil },
				DeleteGroupByIDFunc: func(id string) error { return nil },
			},
			wantOK: true,
		},
		{
			name:   "DeleteGroupByID returns nil but GetGroupByID returns error",
			groups: rootGroup,
			req:    DeleteGroupRequest{ID: "id7"},
			mockRepo: &mockDeleteGroupRepo{
				GetGroupByIDFunc:    func(id string) (acl.Group, error) { return acl.Group{}, errors.New("fail") },
				DeleteGroupByIDFunc: func(id string) error { return nil },
			},
			wantErr: true,
		},
		{
			name:   "forbidden group name edge case",
			groups: rootGroup,
			req:    DeleteGroupRequest{ID: "id8"},
			mockRepo: &mockDeleteGroupRepo{
				GetGroupByIDFunc:    func(id string) (acl.Group, error) { return acl.Group{Name: "forbidden"}, nil },
				DeleteGroupByIDFunc: func(id string) error { return errors.New("forbidden") },
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			if tt.groups != nil {
				ctx = context.WithValue(ctx, model.UserInfoKey, jwtClaimsForGroups(tt.groups))
			}
			uc := DeleteGroupUsecase{repo: tt.mockRepo}
			resp, err := uc.Handle(ctx, tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handle() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantOK && !resp.Success {
				t.Errorf("expected success true")
			}
		})
	}
}
