package entity

import (
	"context"
	"errors"
	"reflect"
	"testing"

	modelentity "github.com/jekiapp/topic-master/internal/model/entity"
)

type mockSaveDescriptionRepo struct {
	getEntityByIDFunc func(id string) (modelentity.Entity, error)
	updateEntityFunc  func(entity modelentity.Entity) error
}

func (m *mockSaveDescriptionRepo) GetEntityByID(id string) (modelentity.Entity, error) {
	return m.getEntityByIDFunc(id)
}
func (m *mockSaveDescriptionRepo) UpdateEntity(entity modelentity.Entity) error {
	return m.updateEntityFunc(entity)
}

func TestSaveDescriptionUsecase_Save(t *testing.T) {
	tests := []struct {
		name         string
		input        SaveDescriptionInput
		mockEntity   modelentity.Entity
		mockGetErr   error
		mockUpdateErr error
		wantResp     SaveDescriptionResponse
		wantErr      bool
	}{
		{
			name: "success",
			input: SaveDescriptionInput{EntityID: "e1", Description: "desc"},
			mockEntity: modelentity.Entity{ID: "e1"},
			wantResp: SaveDescriptionResponse{Message: "Description updated successfully"},
			wantErr: false,
		},
		{
			name: "get entity error",
			input: SaveDescriptionInput{EntityID: "e1", Description: "desc"},
			mockGetErr: errors.New("not found"),
			wantResp: SaveDescriptionResponse{Message: "Failed to get entity"},
			wantErr: true,
		},
		{
			name: "entity not found",
			input: SaveDescriptionInput{EntityID: "e1", Description: "desc"},
			mockEntity: modelentity.Entity{ID: ""},
			wantResp: SaveDescriptionResponse{Message: "Entity not found"},
			wantErr: true,
		},
		{
			name: "update error",
			input: SaveDescriptionInput{EntityID: "e1", Description: "desc"},
			mockEntity: modelentity.Entity{ID: "e1"},
			mockUpdateErr: errors.New("fail"),
			wantResp: SaveDescriptionResponse{Message: "Failed to update description"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &mockSaveDescriptionRepo{
				getEntityByIDFunc: func(id string) (modelentity.Entity, error) {
					if tt.mockGetErr != nil {
						return modelentity.Entity{}, tt.mockGetErr
					}
					return tt.mockEntity, nil
				},
				updateEntityFunc: func(entity modelentity.Entity) error {
					return tt.mockUpdateErr
				},
			}
			uc := SaveDescriptionUsecase{repo: repo}
			got, err := uc.Save(context.Background(), tt.input)
			if !reflect.DeepEqual(got, tt.wantResp) {
				t.Errorf("got = %v, want %v", got, tt.wantResp)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}