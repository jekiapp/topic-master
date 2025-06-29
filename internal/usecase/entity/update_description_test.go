package entity

import (
	"context"
	"errors"
	"testing"
	"github.com/stretchr/testify/assert"

	entitymodel "github.com/jekiapp/topic-master/internal/model/entity"
)

type mockSaveDescriptionRepo struct {
	GetEntityByIDFunc    func(id string) (entitymodel.Entity, error)
	UpdateEntityFunc     func(entitymodel.Entity) error
}

func (m *mockSaveDescriptionRepo) GetEntityByID(id string) (entitymodel.Entity, error) {
	return m.GetEntityByIDFunc(id)
}
func (m *mockSaveDescriptionRepo) UpdateEntity(entity entitymodel.Entity) error {
	return m.UpdateEntityFunc(entity)
}

func TestNewSaveDescriptionUsecase(t *testing.T) {
	uc := NewSaveDescriptionUsecase(nil)
	assert.NotNil(t, uc)
}

func TestSaveDescriptionUsecase_Save(t *testing.T) {
	type fields struct {
		repo iSaveDescriptionRepo
	}
	type args struct {
		ctx   context.Context
		input SaveDescriptionInput
	}

	tests := []struct {
		name      string
		fields    fields
		args      args
		wantMsg   string
		wantErr   bool
	}{
		{
			name: "entity not found",
			fields: fields{repo: &mockSaveDescriptionRepo{
				GetEntityByIDFunc: func(id string) (entitymodel.Entity, error) { return entitymodel.Entity{}, errors.New("not found") },
			}},
			args: args{context.TODO(), SaveDescriptionInput{EntityID: "id", Description: "desc"}},
			wantMsg: "Failed to get entity",
			wantErr: true,
		},
		{
			name: "entity empty id",
			fields: fields{repo: &mockSaveDescriptionRepo{
				GetEntityByIDFunc: func(id string) (entitymodel.Entity, error) { return entitymodel.Entity{}, nil },
			}},
			args: args{context.TODO(), SaveDescriptionInput{EntityID: "id", Description: "desc"}},
			wantMsg: "Entity not found",
			wantErr: true,
		},
		{
			name: "update error",
			fields: fields{repo: &mockSaveDescriptionRepo{
				GetEntityByIDFunc: func(id string) (entitymodel.Entity, error) { return entitymodel.Entity{ID: "id"}, nil },
				UpdateEntityFunc: func(entitymodel.Entity) error { return errors.New("fail update") },
			}},
			args: args{context.TODO(), SaveDescriptionInput{EntityID: "id", Description: "desc"}},
			wantMsg: "Failed to update description",
			wantErr: true,
		},
		{
			name: "success",
			fields: fields{repo: &mockSaveDescriptionRepo{
				GetEntityByIDFunc: func(id string) (entitymodel.Entity, error) { return entitymodel.Entity{ID: "id"}, nil },
				UpdateEntityFunc: func(entitymodel.Entity) error { return nil },
			}},
			args: args{context.TODO(), SaveDescriptionInput{EntityID: "id", Description: "desc"}},
			wantMsg: "Description updated successfully",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := SaveDescriptionUsecase{repo: tt.fields.repo}
			resp, err := uc.Save(tt.args.ctx, tt.args.input)
			assert.Equal(t, tt.wantMsg, resp.Message)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}