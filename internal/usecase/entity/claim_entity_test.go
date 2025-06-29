package entity

import (
	"context"
	"errors"
	"testing"
	"github.com/stretchr/testify/assert"

	acl "github.com/jekiapp/topic-master/internal/model/acl"
	entitymodel "github.com/jekiapp/topic-master/internal/model/entity"
)

type mockClaimEntityRepo struct {
	GetGroupByNameFunc      func(name string) (acl.Group, error)
	GetUserGroupFunc        func(userID, groupID string) (acl.UserGroup, error)
	GetEntityByIDFunc       func(entityID string) (entitymodel.Entity, error)
	CreateApplicationFunc   func(app acl.Application) error
	ListUserGroupsByGroupIDFunc func(groupID string, limit int) ([]acl.UserGroup, error)
	GetReviewerIDsByGroupIDFunc func(groupID string) ([]string, error)
	CreateApplicationAssignmentFunc func(assignment acl.ApplicationAssignment) error
	CreateApplicationHistoryFunc func(history acl.ApplicationHistory) error
}

func (m *mockClaimEntityRepo) GetGroupByName(name string) (acl.Group, error) {
	if m.GetGroupByNameFunc != nil {
		return m.GetGroupByNameFunc(name)
	}
	return acl.Group{}, nil
}
func (m *mockClaimEntityRepo) GetUserGroup(userID, groupID string) (acl.UserGroup, error) {
	if m.GetUserGroupFunc != nil {
		return m.GetUserGroupFunc(userID, groupID)
	}
	return acl.UserGroup{}, nil
}
func (m *mockClaimEntityRepo) GetEntityByID(entityID string) (entitymodel.Entity, error) {
	if m.GetEntityByIDFunc != nil {
		return m.GetEntityByIDFunc(entityID)
	}
	return entitymodel.Entity{}, nil
}
func (m *mockClaimEntityRepo) CreateApplication(app acl.Application) error {
	if m.CreateApplicationFunc != nil {
		return m.CreateApplicationFunc(app)
	}
	return nil
}
func (m *mockClaimEntityRepo) ListUserGroupsByGroupID(groupID string, limit int) ([]acl.UserGroup, error) {
	if m.ListUserGroupsByGroupIDFunc != nil {
		return m.ListUserGroupsByGroupIDFunc(groupID, limit)
	}
	return nil, nil
}
func (m *mockClaimEntityRepo) GetReviewerIDsByGroupID(groupID string) ([]string, error) {
	if m.GetReviewerIDsByGroupIDFunc != nil {
		return m.GetReviewerIDsByGroupIDFunc(groupID)
	}
	return nil, nil
}
func (m *mockClaimEntityRepo) CreateApplicationAssignment(assignment acl.ApplicationAssignment) error {
	if m.CreateApplicationAssignmentFunc != nil {
		return m.CreateApplicationAssignmentFunc(assignment)
	}
	return nil
}
func (m *mockClaimEntityRepo) CreateApplicationHistory(history acl.ApplicationHistory) error {
	if m.CreateApplicationHistoryFunc != nil {
		return m.CreateApplicationHistoryFunc(history)
	}
	return nil
}

func TestNewClaimEntityUsecase(t *testing.T) {
	uc := NewClaimEntityUsecase(nil)
	assert.NotNil(t, uc)
}

func TestClaimEntityRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		input   ClaimEntityRequest
		wantErr bool
	}{
		{"missing entity_id", ClaimEntityRequest{EntityID: "", GroupName: "g"}, true},
		{"missing group_name", ClaimEntityRequest{EntityID: "e", GroupName: ""}, true},
		{"valid", ClaimEntityRequest{EntityID: "e", GroupName: "g"}, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestClaimEntityUsecase_Handle(t *testing.T) {
	type fields struct {
		repo iClaimEntityRepo
	}
	type args struct {
		ctx context.Context
		req ClaimEntityRequest
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "validate error",
			fields: fields{repo: &mockClaimEntityRepo{}},
			args: args{context.TODO(), ClaimEntityRequest{}},
			wantErr: true,
		},
		{
			name: "user not found",
			fields: fields{repo: &mockClaimEntityRepo{}},
			args: args{context.Background(), ClaimEntityRequest{EntityID: "e", GroupName: "g"}},
			wantErr: true,
		},
		{
			name: "group not found",
			fields: fields{repo: &mockClaimEntityRepo{
				GetGroupByNameFunc: func(name string) (acl.Group, error) { return acl.Group{}, errors.New("not found") },
			}},
			args: args{context.WithValue(context.Background(), "user", &struct{ID string}{ID: "u1"}), ClaimEntityRequest{EntityID: "e", GroupName: "g"}},
			wantErr: true,
		},
		{
			name: "user not in group",
			fields: fields{repo: &mockClaimEntityRepo{
				GetGroupByNameFunc: func(name string) (acl.Group, error) { return acl.Group{ID: "g"}, nil },
				GetUserGroupFunc: func(userID, groupID string) (acl.UserGroup, error) { return acl.UserGroup{}, errors.New("not member") },
			}},
			args: args{context.WithValue(context.Background(), "user", &struct{ID string}{ID: "u1"}), ClaimEntityRequest{EntityID: "e", GroupName: "g"}},
			wantErr: true,
		},
		{
			name: "entity not found",
			fields: fields{repo: &mockClaimEntityRepo{
				GetGroupByNameFunc: func(name string) (acl.Group, error) { return acl.Group{ID: "g"}, nil },
				GetUserGroupFunc: func(userID, groupID string) (acl.UserGroup, error) { return acl.UserGroup{}, nil },
				GetEntityByIDFunc: func(entityID string) (entitymodel.Entity, error) { return entitymodel.Entity{}, errors.New("not found") },
			}},
			args: args{context.WithValue(context.Background(), "user", &struct{ID string}{ID: "u1"}), ClaimEntityRequest{EntityID: "e", GroupName: "g"}},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := ClaimEntityUsecase{repo: tt.fields.repo}
			_, err := uc.Handle(tt.args.ctx, tt.args.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}