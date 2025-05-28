// I want to implement this usecase to list all the applications that assigned to me
// get the user id from the context
// learn from get_group_list.go for the pattern
// get the assignment by reviewer id see application.go

package application

import (
	"context"

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
	return dbpkg.SelectAll[acl.ApplicationAssignment](r.db, reviewerID, acl.IdxAppAssign_ReviewerID)
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

func (uc ListMyAssignmentUsecase) Handle(ctx context.Context, req ListMyAssignmentRequest) (ListMyAssignmentResponse, error) {
	user := util.GetUserInfo(ctx)
	assignments, err := uc.repo.ListAssignmentsByReviewerID(user.ID)
	if err != nil {
		return ListMyAssignmentResponse{}, err
	}
	var apps []acl.Application
	for _, assign := range assignments {
		app, err := uc.repo.GetApplicationByID(assign.ApplicationID)
		if err == nil {
			apps = append(apps, app)
		}
	}
	return ListMyAssignmentResponse{Applications: apps}, nil
}
