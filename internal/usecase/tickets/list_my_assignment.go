// I want to implement this usecase to list all the applications that assigned to me
// get the user id from the context
// learn from get_group_list.go for the pattern
// get the assignment by reviewer id see application.go

package tickets

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/jekiapp/topic-master/internal/model/acl"
	dbpkg "github.com/jekiapp/topic-master/pkg/db"
	"github.com/jekiapp/topic-master/pkg/util"
	"github.com/tidwall/buntdb"
)

type ListMyAssignmentRequest struct{}

type ListMyAssignmentResponse struct {
	Applications []acl.Application `json:"applications"`
}

type iAssignmentRepo interface {
	ListAssignmentsByReviewerID(reviewerID string) ([]acl.ApplicationAssignment, error)
	GetApplicationByID(appID string) (acl.Application, error)
}

type assignmentRepoImpl struct {
	db *buntdb.DB
}

func (r *assignmentRepoImpl) ListAssignmentsByReviewerID(reviewerID string) ([]acl.ApplicationAssignment, error) {
	pivot := fmt.Sprintf("%s:%d", reviewerID, time.Now().Unix())
	return dbpkg.SelectAll[acl.ApplicationAssignment](r.db, "-<="+pivot, acl.IdxAppAssign_ReviewerID)
}

func (r *assignmentRepoImpl) GetApplicationByID(appID string) (acl.Application, error) {
	return dbpkg.GetByID[acl.Application](r.db, appID)
}

type ListMyAssignmentUsecase struct {
	repo iAssignmentRepo
}

func NewListMyAssignmentUsecase(db *buntdb.DB) ListMyAssignmentUsecase {
	return ListMyAssignmentUsecase{
		repo: &assignmentRepoImpl{db: db},
	}
}

func (uc ListMyAssignmentUsecase) Handle(ctx context.Context, req map[string]string) (ListMyAssignmentResponse, error) {
	user := util.GetUserInfo(ctx)
	if user == nil {
		return ListMyAssignmentResponse{}, fmt.Errorf("unauthorized: user info not found")
	}
	assignments, err := uc.repo.ListAssignmentsByReviewerID(user.ID)
	if err != nil && err != dbpkg.ErrNotFound {
		return ListMyAssignmentResponse{}, err
	}
	var apps []acl.Application
	for _, assign := range assignments {
		app, err := uc.repo.GetApplicationByID(assign.ApplicationID)
		if err == nil {
			apps = append(apps, app)
		} else {
			log.Println("error getting application by id", err)
		}
	}
	return ListMyAssignmentResponse{Applications: apps}, nil
}
