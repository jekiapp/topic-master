package group

import (
	"context"
	"errors"
	"testing"

	"github.com/jekiapp/topic-master/internal/model/acl"
)

type mockGroupListRepo struct {
	GetAllGroupsFunc func() ([]acl.Group, error)
}

func (m *mockGroupListRepo) GetAllGroups() ([]acl.Group, error) {
	return m.GetAllGroupsFunc()
}

type mockUserGroupRepo struct {
	ListUserGroupsByGroupIDFunc func(groupID string, limit int) ([]acl.UserGroup, error)
	GetUserByIDFunc             func(userID string) (acl.User, error)
}

func (m *mockUserGroupRepo) ListUserGroupsByGroupID(groupID string, limit int) ([]acl.UserGroup, error) {
	return m.ListUserGroupsByGroupIDFunc(groupID, limit)
}
func (m *mockUserGroupRepo) GetUserByID(userID string) (acl.User, error) {
	return m.GetUserByIDFunc(userID)
}

func TestGetGroupListUsecase_Handle(t *testing.T) {
	tests := []struct {
		name         string
		mockGroup    *mockGroupListRepo
		mockUserRepo *mockUserGroupRepo
		wantErr      bool
		wantLen      int
	}{
		{
			name: "repo error",
			mockGroup: &mockGroupListRepo{
				GetAllGroupsFunc: func() ([]acl.Group, error) { return nil, errors.New("fail") },
			},
			mockUserRepo: &mockUserGroupRepo{},
			wantErr:      true,
		},
		{
			name: "success with members",
			mockGroup: &mockGroupListRepo{
				GetAllGroupsFunc: func() ([]acl.Group, error) {
					return []acl.Group{{ID: "g1", Name: "n1", Description: "d1"}}, nil
				},
			},
			mockUserRepo: &mockUserGroupRepo{
				ListUserGroupsByGroupIDFunc: func(groupID string, limit int) ([]acl.UserGroup, error) {
					return []acl.UserGroup{{UserID: "u1"}}, nil
				},
				GetUserByIDFunc: func(userID string) (acl.User, error) {
					return acl.User{Username: "user1", Status: acl.StatusUserActive}, nil
				},
			},
			wantLen: 1,
		},
		{
			name: "empty group list",
			mockGroup: &mockGroupListRepo{
				GetAllGroupsFunc: func() ([]acl.Group, error) { return nil, nil },
			},
			mockUserRepo: &mockUserGroupRepo{},
			wantLen:      0,
		},
		{
			name: "user repo returns error",
			mockGroup: &mockGroupListRepo{
				GetAllGroupsFunc: func() ([]acl.Group, error) {
					return []acl.Group{{ID: "g2", Name: "n2", Description: "d2"}}, nil
				},
			},
			mockUserRepo: &mockUserGroupRepo{
				ListUserGroupsByGroupIDFunc: func(groupID string, limit int) ([]acl.UserGroup, error) {
					return nil, errors.New("fail")
				},
			},
			wantLen: 1,
		},
		{
			name: "group with no active users",
			mockGroup: &mockGroupListRepo{
				GetAllGroupsFunc: func() ([]acl.Group, error) {
					return []acl.Group{{ID: "g3", Name: "n3", Description: "d3"}}, nil
				},
			},
			mockUserRepo: &mockUserGroupRepo{
				ListUserGroupsByGroupIDFunc: func(groupID string, limit int) ([]acl.UserGroup, error) {
					return []acl.UserGroup{{UserID: "u2"}}, nil
				},
				GetUserByIDFunc: func(userID string) (acl.User, error) {
					return acl.User{Username: "user2", Status: "inactive"}, nil
				},
			},
			wantLen: 1,
		},
		{
			name: "group with multiple users, some inactive",
			mockGroup: &mockGroupListRepo{
				GetAllGroupsFunc: func() ([]acl.Group, error) {
					return []acl.Group{{ID: "g4", Name: "n4", Description: "d4"}}, nil
				},
			},
			mockUserRepo: &mockUserGroupRepo{
				ListUserGroupsByGroupIDFunc: func(groupID string, limit int) ([]acl.UserGroup, error) {
					return []acl.UserGroup{{UserID: "u3"}, {UserID: "u4"}}, nil
				},
				GetUserByIDFunc: func(userID string) (acl.User, error) {
					if userID == "u3" {
						return acl.User{Username: "user3", Status: acl.StatusUserActive}, nil
					}
					return acl.User{Username: "user4", Status: "inactive"}, nil
				},
			},
			wantLen: 1,
		},
		{
			name: "multiple groups",
			mockGroup: &mockGroupListRepo{
				GetAllGroupsFunc: func() ([]acl.Group, error) {
					return []acl.Group{
						{ID: "g1", Name: "n1", Description: "d1"},
						{ID: "g2", Name: "n2", Description: "d2"},
					}, nil
				},
			},
			mockUserRepo: &mockUserGroupRepo{
				ListUserGroupsByGroupIDFunc: func(groupID string, limit int) ([]acl.UserGroup, error) {
					return []acl.UserGroup{{UserID: "u1"}}, nil
				},
				GetUserByIDFunc: func(userID string) (acl.User, error) {
					return acl.User{Username: "user1", Status: acl.StatusUserActive}, nil
				},
			},
			wantLen: 2,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := GetGroupListUsecase{groupRepo: tt.mockGroup, userGroupRepo: tt.mockUserRepo}
			resp, err := uc.Handle(context.Background(), GetGroupListRequest{})
			if (err != nil) != tt.wantErr {
				t.Errorf("Handle() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(resp.Groups) != tt.wantLen {
				t.Errorf("expected %d groups, got %d", tt.wantLen, len(resp.Groups))
			}
			if tt.wantLen > 0 && resp.Groups[0].Members != "user1" {
				t.Errorf("expected members 'user1', got '%s'", resp.Groups[0].Members)
			}
		})
	}
}
