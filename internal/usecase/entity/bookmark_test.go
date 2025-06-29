package entity

import (
	"context"
	"errors"
	"reflect"
	"testing"

	entitymodel "github.com/jekiapp/topic-master/internal/model/entity"
	"github.com/tidwall/buntdb"
)

type mockBookmarkRepo struct {
	toggleBookmarkFunc   func(entityID, entityType, userID string, bookmark bool) error
	getEntityByIDFunc    func(db *buntdb.DB, entityID string) (entitymodel.Entity, error)
}

func (m *mockBookmarkRepo) ToggleBookmark(entityID, entityType, userID string, bookmark bool) error {
	return m.toggleBookmarkFunc(entityID, entityType, userID, bookmark)
}
func (m *mockBookmarkRepo) GetEntityByID(db *buntdb.DB, entityID string) (entitymodel.Entity, error) {
	return m.getEntityByIDFunc(db, entityID)
}

// Custom context key type to avoid collisions
type userInfoKeyType struct{}
var userInfoKey = userInfoKeyType{}

func TestToggleBookmarkUsecase_Toggle(t *testing.T) {
	tests := []struct {
		name       string
		ctx        context.Context
		input      ToggleBookmarkInput
		mockUserID string
		mockEntity entitymodel.Entity
		mockGetErr error
		mockToggleErr error
		wantResp   ToggleBookmarkResponse
		wantErr    bool
	}{
		{
			name: "success",
			mockUserID: "user1",
			mockEntity: entitymodel.Entity{ID: "e1", TypeID: "type1"},
			ctx: context.WithValue(context.Background(), userInfoKey, &struct{ID string}{ID: "user1"}),
			input: ToggleBookmarkInput{EntityID: "e1", Bookmark: true},
			wantResp: ToggleBookmarkResponse{Message: "Bookmark toggled successfully"},
			wantErr: false,
		},
		{
			name: "user info missing",
			ctx: context.Background(),
			input: ToggleBookmarkInput{EntityID: "e1", Bookmark: true},
			wantResp: ToggleBookmarkResponse{Message: "User info not found in context"},
			wantErr: true,
		},
		{
			name: "entity not found",
			mockUserID: "user1",
			mockGetErr: errors.New("not found"),
			ctx: context.WithValue(context.Background(), userInfoKey, &struct{ID string}{ID: "user1"}),
			input: ToggleBookmarkInput{EntityID: "e1", Bookmark: true},
			wantResp: ToggleBookmarkResponse{Message: "Entity not found"},
			wantErr: true,
		},
		{
			name: "already exists",
			mockUserID: "user1",
			mockEntity: entitymodel.Entity{ID: "e1", TypeID: "type1"},
			mockToggleErr: errors.New("already exists"),
			ctx: context.WithValue(context.Background(), userInfoKey, &struct{ID string}{ID: "user1"}),
			input: ToggleBookmarkInput{EntityID: "e1", Bookmark: true},
			wantResp: ToggleBookmarkResponse{Message: "Bookmark already exists"},
			wantErr: false,
		},
		{
			name: "toggle error",
			mockUserID: "user1",
			mockEntity: entitymodel.Entity{ID: "e1", TypeID: "type1"},
			mockToggleErr: errors.New("fail"),
			ctx: context.WithValue(context.Background(), userInfoKey, &struct{ID string}{ID: "user1"}),
			input: ToggleBookmarkInput{EntityID: "e1", Bookmark: true},
			wantResp: ToggleBookmarkResponse{Message: "Failed to toggle bookmark"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockBookmarkRepo{
				getEntityByIDFunc: func(db *buntdb.DB, entityID string) (entitymodel.Entity, error) {
					if tt.mockGetErr != nil {
						return entitymodel.Entity{}, tt.mockGetErr
					}
					return tt.mockEntity, nil
				},
				toggleBookmarkFunc: func(entityID, entityType, userID string, bookmark bool) error {
					return tt.mockToggleErr
				},
			}
			uc := ToggleBookmarkUsecase{repo: repo, db: nil}
			got, err := uc.Toggle(tt.ctx, tt.input)
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("got = %v, want %v", got, tt.wantResp)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}