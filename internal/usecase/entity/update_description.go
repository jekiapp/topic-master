// this usecase is to save description for entity field
// it will receive entity id and description string
// the userid should be get from the context
//

package entity

import (
	"context"
	"errors"
	"time"

	"github.com/jekiapp/topic-master/internal/model/entity"
	dbpkg "github.com/jekiapp/topic-master/pkg/db"
	"github.com/jekiapp/topic-master/pkg/util"
	"github.com/tidwall/buntdb"
)

// SaveDescriptionInput holds the parameters for saving the description for an entity.
type SaveDescriptionInput struct {
	EntityID    string
	Description string
}

type SaveDescriptionResponse struct {
	Message string `json:"message"`
}

// iSaveDescriptionRepo abstracts the repository for entity persistence.
type iSaveDescriptionRepo interface {
	GetEntityByID(id string) (entity.Entity, error)
	UpdateEntity(entity entity.Entity) error
}

// saveDescriptionRepo implements iSaveDescriptionRepo using buntdb.
type saveDescriptionRepo struct {
	db *buntdb.DB
}

func (r *saveDescriptionRepo) GetEntityByID(id string) (entity.Entity, error) {
	return getEntityByID(r.db, id)
}

func (r *saveDescriptionRepo) UpdateEntity(entity entity.Entity) error {
	return updateEntity(r.db, entity)
}

// SaveDescriptionUsecase handles saving the description for an entity.
type SaveDescriptionUsecase struct {
	repo iSaveDescriptionRepo
}

func NewSaveDescriptionUsecase(db *buntdb.DB) SaveDescriptionUsecase {
	return SaveDescriptionUsecase{
		repo: &saveDescriptionRepo{db: db},
	}
}

// Save updates the description for the given entity, extracting user ID from context.
func (uc SaveDescriptionUsecase) Save(ctx context.Context, input SaveDescriptionInput) (SaveDescriptionResponse, error) {
	entity, err := uc.repo.GetEntityByID(input.EntityID)
	if err != nil {
		return SaveDescriptionResponse{Message: "Failed to get entity"}, err
	}
	if entity.ID == "" {
		return SaveDescriptionResponse{Message: "Entity not found"}, errors.New("entity not found")
	}
	entity.Description = input.Description
	entity.UpdatedAt = time.Now()

	updatedBy := "Anonymous"
	userInfo := util.GetUserInfo(ctx)
	if userInfo != nil {
		updatedBy = userInfo.ID
	}

	entity.Metadata["updated_by"] = updatedBy
	err = uc.repo.UpdateEntity(entity)
	if err != nil {
		return SaveDescriptionResponse{Message: "Failed to update description"}, err
	}
	return SaveDescriptionResponse{Message: "Description updated successfully"}, nil
}

// --- helpers ---

// getEntityByID fetches an entity by ID using the repository pattern.
func getEntityByID(db *buntdb.DB, id string) (entity.Entity, error) {
	entityObj, err := dbpkg.GetByID[entity.Entity](db, id)
	if err != nil {
		return entity.Entity{}, err
	}
	return entityObj, nil
}

// updateEntity persists the entity using db.Update.
func updateEntity(db *buntdb.DB, entity entity.Entity) error {
	return dbpkg.Update(db, entity)
}
