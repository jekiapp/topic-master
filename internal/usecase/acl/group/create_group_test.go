package group

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/jekiapp/topic-master/internal/model/acl"
)

type mockCreateGroupRepo struct {
	CreateGroupFunc    func(group acl.Group) error
	GetGroupByNameFunc func(name string) (acl.Group, error)
}

func (m *mockCreateGroupRepo) CreateGroup(group acl.Group) error {
	return m.CreateGroupFunc(group)
}
func (m *mockCreateGroupRepo) GetGroupByName(name string) (acl.Group, error) {
	return m.GetGroupByNameFunc(name)
}

func TestCreateGroupUsecase_Handle(t *testing.T) {
	tests := []struct {
		name      string
		req       CreateGroupRequest
		mockRepo  *mockCreateGroupRepo
		wantErr   bool
		wantGroup bool
	}{
		{
			name:     "missing name",
			req:      CreateGroupRequest{},
			mockRepo: &mockCreateGroupRepo{},
			wantErr:  true,
		},
		{
			name: "group already exists",
			req:  CreateGroupRequest{Name: "group1"},
			mockRepo: &mockCreateGroupRepo{
				GetGroupByNameFunc: func(name string) (acl.Group, error) { return acl.Group{}, nil },
			},
			wantErr: true,
		},
		{
			name: "repo create error",
			req:  CreateGroupRequest{Name: "group2"},
			mockRepo: &mockCreateGroupRepo{
				GetGroupByNameFunc: func(name string) (acl.Group, error) { return acl.Group{}, errors.New("not found") },
				CreateGroupFunc:    func(group acl.Group) error { return errors.New("fail") },
			},
			wantErr: true,
		},
		{
			name: "success",
			req:  CreateGroupRequest{Name: "group3", Description: "desc"},
			mockRepo: &mockCreateGroupRepo{
				GetGroupByNameFunc: func(name string) (acl.Group, error) { return acl.Group{}, errors.New("not found") },
				CreateGroupFunc:    func(group acl.Group) error { return nil },
			},
			wantErr:   false,
			wantGroup: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := CreateGroupUsecase{repo: tt.mockRepo}
			resp, err := uc.Handle(context.Background(), tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handle() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantGroup && reflect.DeepEqual(resp.Group, acl.Group{}) {
				t.Errorf("expected group in response")
			}
		})
	}
}
