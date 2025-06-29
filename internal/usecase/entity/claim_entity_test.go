package entity

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/jekiapp/topic-master/internal/model/acl"
	entitymodel "github.com/jekiapp/topic-master/internal/model/entity"
)

type mockClaimEntityRepo struct {
	getGroupByNameFunc      func(name string) (acl.Group, error)
	getUserGroupFunc        func(userID, groupID string) (acl.UserGroup, error)
	getEntityByIDFunc       func(entityID string) (entitymodel.Entity, error)
	createApplicationFunc   func(app acl.Application) error
}

func (m *mockClaimEntityRepo) CreateApplication(app acl.Application) error {
	return m.createApplicationFunc(app)
}
func (m *mockClaimEntityRepo) GetGroupByName(name string) (acl.Group, error) {
	return m.getGroupByNameFunc(name)
}
func (m *mockClaimEntityRepo) GetUserGroup(userID, groupID string) (acl.UserGroup, error) {
	return m.getUserGroupFunc(userID, groupID)
}
func (m *mockClaimEntityRepo) ListUserGroupsByGroupID(groupID string, limit int) ([]acl.UserGroup, error) {
	return nil, nil
}
func (m *mockClaimEntityRepo) GetReviewerIDsByGroupID(groupID string) ([]string, error) {
	return nil, nil
}
func (m *mockClaimEntityRepo) CreateApplicationAssignment(assignment acl.ApplicationAssignment) error {
	return nil
}
func (m *mockClaimEntityRepo) CreateApplicationHistory(history acl.ApplicationHistory) error {
	return nil
}
func (m *mockClaimEntityRepo) GetEntityByID(entityID string) (entitymodel.Entity, error) {
	return m.getEntityByIDFunc(entityID)
}

func TestClaimEntityRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		input   ClaimEntityRequest
		wantErr bool
	}{
		{"ok", ClaimEntityRequest{EntityID: "e", GroupName: "g"}, false},
		{"missing entity_id", ClaimEntityRequest{EntityID: "", GroupName: "g"}, true},
		{"missing group_name", ClaimEntityRequest{EntityID: "e", GroupName: ""}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.input.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("got error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestClaimEntityUsecase_Handle(t *testing.T) {
	tests := []struct {
		name         string
		input        ClaimEntityRequest
		mockUser     *struct{ID string}
		mockGroup    acl.Group
		mockEntity   entitymodel.Entity
		mockValidateErr error
		mockGetGroupErr error
		mockGetUserGroupErr error
		mockGetEntityErr error
		mockCreateAppErr error
		wantResp     ClaimEntityResponse
		wantErr      bool
	}{
		{
			name: "success",
			input: ClaimEntityRequest{EntityID: "e1", GroupName: "g1", Reason: "r"},
			mockUser: &struct{ID string}{ID: "u1"},
			mockGroup: acl.Group{ID: "g1"},
			mockEntity: entitymodel.Entity{ID: "e1", TypeID: "type", Name: "n"},
			wantResp: ClaimEntityResponse{ApplicationID: "app1", LinkRedirect: "/#ticket-detail?id=app1"},
			wantErr: false,
		},
		{
			name: "validate error",
			input: ClaimEntityRequest{EntityID: "", GroupName: "g1"},
			mockValidateErr: errors.New("missing entity_id"),
			wantResp: ClaimEntityResponse{},
			wantErr: true,
		},
		{
			name: "user not found",
			input: ClaimEntityRequest{EntityID: "e1", GroupName: "g1"},
			mockUser: nil,
			wantResp: ClaimEntityResponse{},
			wantErr: true,
		},
		{
			name: "group not found",
			input: ClaimEntityRequest{EntityID: "e1", GroupName: "g1"},
			mockUser: &struct{ID string}{ID: "u1"},
			mockGetGroupErr: errors.New("not found"),
			wantResp: ClaimEntityResponse{},
			wantErr: true,
		},
		{
			name: "user not in group",
			input: ClaimEntityRequest{EntityID: "e1", GroupName: "g1"},
			mockUser: &struct{ID string}{ID: "u1"},
			mockGroup: acl.Group{ID: "g1"},
			mockGetUserGroupErr: errors.New("not member"),
			wantResp: ClaimEntityResponse{},
			wantErr: true,
		},
		{
			name: "entity not found",
			input: ClaimEntityRequest{EntityID: "e1", GroupName: "g1"},
			mockUser: &struct{ID string}{ID: "u1"},
			mockGroup: acl.Group{ID: "g1"},
			mockEntity: entitymodel.Entity{},
			mockGetEntityErr: errors.New("not found"),
			wantResp: ClaimEntityResponse{},
			wantErr: true,
		},
		{
			name: "create app error",
			input: ClaimEntityRequest{EntityID: "e1", GroupName: "g1", Reason: "r"},
			mockUser: &struct{ID string}{ID: "u1"},
			mockGroup: acl.Group{ID: "g1"},
			mockEntity: entitymodel.Entity{ID: "e1", TypeID: "type", Name: "n"},
			mockCreateAppErr: errors.New("fail"),
			wantResp: ClaimEntityResponse{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockClaimEntityRepo{
				getGroupByNameFunc: func(name string) (acl.Group, error) {
					if tt.mockGetGroupErr != nil {
						return acl.Group{}, tt.mockGetGroupErr
					}
					return tt.mockGroup, nil
				},
				getUserGroupFunc: func(userID, groupID string) (acl.UserGroup, error) {
					if tt.mockGetUserGroupErr != nil {
						return acl.UserGroup{}, tt.mockGetUserGroupErr
					}
					return acl.UserGroup{}, nil
				},
				getEntityByIDFunc: func(entityID string) (entitymodel.Entity, error) {
					if tt.mockGetEntityErr != nil {
						return entitymodel.Entity{}, tt.mockGetEntityErr
					}
					return tt.mockEntity, nil
				},
				createApplicationFunc: func(app acl.Application) error {
					if tt.mockCreateAppErr != nil {
						return tt.mockCreateAppErr
					}
					return nil
				},
			}
			uc := ClaimEntityUsecase{repo: repo}
			ctx := context.Background()
			if tt.mockUser != nil {
				ctx = context.WithValue(ctx, userInfoKey, tt.mockUser)
			}
			got, err := uc.Handle(ctx, tt.input)
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("got = %v, want %v", got, tt.wantResp)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}