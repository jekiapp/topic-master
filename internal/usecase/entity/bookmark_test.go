package entity

import (
	"context"
	"errors"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/buntdb"

	entitymodel "github.com/jekiapp/topic-master/internal/model/entity"
	acl "github.com/jekiapp/topic-master/internal/model/acl"
	"github.com/jekiapp/topic-master/internal/model"
)

type mockBookmarkRepo struct {
	ToggleBookmarkFunc func(entityID, entityType, userID string, bookmark bool) error
	GetEntityByIDFunc  func(db *buntdb.DB, entityID string) (entitymodel.Entity, error)
}

func (m *mockBookmarkRepo) ToggleBookmark(entityID, entityType, userID string, bookmark bool) error {
	return m.ToggleBookmarkFunc(entityID, entityType, userID, bookmark)
}
func (m *mockBookmarkRepo) GetEntityByID(db *buntdb.DB, entityID string) (entitymodel.Entity, error) {
	return m.GetEntityByIDFunc(db, entityID)
}

func TestNewToggleBookmarkUsecase(t *testing.T) {
	uc := NewToggleBookmarkUsecase(nil)
	assert.NotNil(t, uc)
}

func TestToggleBookmarkUsecase_Toggle(t *testing.T) {
	db := &buntdb.DB{} // dummy db for test
	type fields struct {
		repo iBookmarkRepo
		db   *buntdb.DB
	}
	type args struct {
		ctx   context.Context
		input ToggleBookmarkInput
	}

	userCtx := func() context.Context {
		return context.WithValue(context.Background(), model.UserInfoKey, &acl.JWTClaims{UserID: "u1"})
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantMsg string
		wantErr bool
	}{
		{
			name: "user info not found",
			fields: fields{repo: &mockBookmarkRepo{}, db: db},
			args: args{context.Background(), ToggleBookmarkInput{EntityID: "id", Bookmark: true}},
			wantMsg: "User info not found in context",
			wantErr: true,
		},
		{
			name: "entity not found",
			fields: fields{repo: &mockBookmarkRepo{
				GetEntityByIDFunc: func(db *buntdb.DB, entityID string) (entitymodel.Entity, error) { return entitymodel.Entity{}, errors.New("not found") },
			}, db: db},
			args: args{userCtx(), ToggleBookmarkInput{EntityID: "id", Bookmark: true}},
			wantMsg: "Entity not found",
			wantErr: true,
		},
		{
			name: "toggle error already exists",
			fields: fields{repo: &mockBookmarkRepo{
				GetEntityByIDFunc: func(db *buntdb.DB, entityID string) (entitymodel.Entity, error) { return entitymodel.Entity{ID: "id"}, nil },
				ToggleBookmarkFunc: func(entityID, entityType, userID string, bookmark bool) error { return errors.New("already exists") },
			}, db: db},
			args: args{userCtx(), ToggleBookmarkInput{EntityID: "id", Bookmark: true}},
			wantMsg: "Bookmark already exists",
			wantErr: false,
		},
		{
			name: "toggle error other",
			fields: fields{repo: &mockBookmarkRepo{
				GetEntityByIDFunc: func(db *buntdb.DB, entityID string) (entitymodel.Entity, error) { return entitymodel.Entity{ID: "id"}, nil },
				ToggleBookmarkFunc: func(entityID, entityType, userID string, bookmark bool) error { return errors.New("fail toggle") },
			}, db: db},
			args: args{userCtx(), ToggleBookmarkInput{EntityID: "id", Bookmark: true}},
			wantMsg: "Failed to toggle bookmark",
			wantErr: true,
		},
		{
			name: "success",
			fields: fields{repo: &mockBookmarkRepo{
				GetEntityByIDFunc: func(db *buntdb.DB, entityID string) (entitymodel.Entity, error) { return entitymodel.Entity{ID: "id", TypeID: "type"}, nil },
				ToggleBookmarkFunc: func(entityID, entityType, userID string, bookmark bool) error { return nil },
			}, db: db},
			args: args{userCtx(), ToggleBookmarkInput{EntityID: "id", Bookmark: true}},
			wantMsg: "Bookmark toggled successfully",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := ToggleBookmarkUsecase{repo: tt.fields.repo, db: tt.fields.db}
			resp, err := uc.Toggle(tt.args.ctx, tt.args.input)
			assert.Equal(t, tt.wantMsg, resp.Message)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}