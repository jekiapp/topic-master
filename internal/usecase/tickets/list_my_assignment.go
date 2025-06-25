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

type ListMyAssignmentResponse struct {
	Applications []applicationResponse `json:"applications"`
	HasNext      bool                  `json:"has_next"`
}

type applicationResponse struct {
	acl.Application
	ApplicantName string `json:"applicant_name"`
}

type iAssignmentRepo interface {
	ListAssignmentsByReviewerIDPaginated(reviewerID string, page, limit int) ([]acl.ApplicationAssignment, bool, error)
	GetApplicationByID(appID string) (acl.Application, error)
	GetUserByID(userID string) (acl.User, error)
}

type assignmentRepoImpl struct {
	db *buntdb.DB
}

func (r *assignmentRepoImpl) ListAssignmentsByReviewerIDPaginated(reviewerID string, page, limit int) ([]acl.ApplicationAssignment, bool, error) {
	pivot := fmt.Sprintf("%s:%d", reviewerID, time.Now().Unix())
	assignments, err := dbpkg.SelectPaginated[acl.ApplicationAssignment](r.db, "-<="+pivot, acl.IdxAppAssign_ReviewerID, &dbpkg.Pagination{Page: page, Limit: limit + 1})
	if err != nil && err != dbpkg.ErrNotFound {
		return nil, false, err
	}
	hasNext := false
	if len(assignments) > limit {
		hasNext = true
		assignments = assignments[:limit]
	}
	return assignments, hasNext, nil
}

func (r *assignmentRepoImpl) GetApplicationByID(appID string) (acl.Application, error) {
	return dbpkg.GetByID[acl.Application](r.db, appID)
}

func (r *assignmentRepoImpl) GetUserByID(userID string) (acl.User, error) {
	return dbpkg.GetByID[acl.User](r.db, userID)
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
	// parse page and limit from req
	page, limit := 1, 20
	if v, ok := req["page"]; ok {
		fmt.Sscanf(v, "%d", &page)
	}
	if v, ok := req["limit"]; ok {
		fmt.Sscanf(v, "%d", &limit)
	}
	assignments, hasNext, err := uc.repo.ListAssignmentsByReviewerIDPaginated(user.ID, page, limit)
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
	appsResp := []applicationResponse{}
	for _, app := range apps {
		user, err := uc.repo.GetUserByID(app.UserID)
		name := ""
		if err == nil {
			name = user.Username
		}
		appsResp = append(appsResp, applicationResponse{
			Application:   app,
			ApplicantName: name,
		})
	}
	return ListMyAssignmentResponse{Applications: appsResp, HasNext: hasNext}, nil
}
