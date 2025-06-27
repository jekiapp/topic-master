// in this usecase we will show the detail of the signup application
// it will show the results from signup.go

package acl

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jekiapp/topic-master/internal/model/acl"
	userrepo "github.com/jekiapp/topic-master/internal/repository/user"
	dbpkg "github.com/jekiapp/topic-master/pkg/db"
	"github.com/tidwall/buntdb"
)

// Request struct for viewing signup application
// Only needs ApplicationID

type ViewSignupApplicationRequest struct {
	ApplicationID string `json:"application_id"`
}

type ViewSignupApplicationResponse struct {
	Application acl.Application          `json:"application"`
	User        acl.User                 `json:"user"`
	Assignee    []Assignee               `json:"assignee"`
	Histories   []acl.ApplicationHistory `json:"histories"`
}

type Assignee struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Status   string `json:"status"`
}

type IViewSignupApplicationRepo interface {
	GetApplicationByID(id string) (acl.Application, error)
	GetUserByID(id string) (acl.User, error)
	GetUserPendingByID(id string) (acl.UserPending, error)
	GetGroupByID(id string) (acl.Group, error)
	GetUserGroup(userID, groupID string) (acl.UserGroup, error)
	ListAssignmentsByApplicationID(appID string) ([]acl.ApplicationAssignment, error)
	ListHistoriesByApplicationID(appID string) ([]acl.ApplicationHistory, error)
}

type viewSignupApplicationRepo struct {
	db *buntdb.DB
}

func (r *viewSignupApplicationRepo) GetApplicationByID(id string) (acl.Application, error) {
	return dbpkg.GetByID[acl.Application](r.db, id)
}

func (r *viewSignupApplicationRepo) GetUserByID(id string) (acl.User, error) {
	return dbpkg.GetByID[acl.User](r.db, id)
}

func (r *viewSignupApplicationRepo) GetUserPendingByID(id string) (acl.UserPending, error) {
	return dbpkg.GetByID[acl.UserPending](r.db, id)
}

func (r *viewSignupApplicationRepo) GetGroupByID(id string) (acl.Group, error) {
	return userrepo.GetGroupByID(r.db, id)
}

func (r *viewSignupApplicationRepo) GetUserGroup(userID, groupID string) (acl.UserGroup, error) {
	return userrepo.GetUserGroup(r.db, userID, groupID)
}

func (r *viewSignupApplicationRepo) ListAssignmentsByApplicationID(appID string) ([]acl.ApplicationAssignment, error) {
	return dbpkg.SelectAll[acl.ApplicationAssignment](r.db, "="+appID, acl.IdxAppAssign_ApplicationID)
}

func (r *viewSignupApplicationRepo) ListHistoriesByApplicationID(appID string) ([]acl.ApplicationHistory, error) {
	appID = fmt.Sprintf("%s:%d", appID, time.Now().Unix())
	return dbpkg.SelectAll[acl.ApplicationHistory](r.db, "-<="+appID, acl.IdxAppHistory_ApplicationID)
}

type ViewSignupApplicationUsecase struct {
	repo IViewSignupApplicationRepo
}

func NewViewSignupApplicationUsecase(db *buntdb.DB) ViewSignupApplicationUsecase {
	return ViewSignupApplicationUsecase{
		repo: &viewSignupApplicationRepo{db: db},
	}
}

func (uc ViewSignupApplicationUsecase) Handle(ctx context.Context, req map[string]string) (ViewSignupApplicationResponse, error) {
	appID := req["id"]
	app, err := uc.repo.GetApplicationByID(appID)
	if err != nil {
		return ViewSignupApplicationResponse{}, errors.New("application not found")
	}
	user, err := uc.repo.GetUserPendingByID(app.UserID)
	if err != nil {
		return ViewSignupApplicationResponse{}, errors.New("user not found")
	}

	assignments, err := uc.repo.ListAssignmentsByApplicationID(appID)
	if err != nil {
		return ViewSignupApplicationResponse{}, fmt.Errorf("assignments not found: %w", err)
	}
	assignees := []Assignee{}
	for _, assignment := range assignments {
		user, err := uc.repo.GetUserByID(assignment.ReviewerID)
		if err != nil {
			log.Println("[SIGNUP APPLICATION] error getting username by user id", err)
		}
		assignees = append(assignees, Assignee{
			UserID:   user.ID,
			Username: user.Username,
			Name:     user.Name,
			Status:   assignment.ReviewStatus,
		})
	}

	histories, err := uc.repo.ListHistoriesByApplicationID(appID)
	if err != nil {
		return ViewSignupApplicationResponse{}, fmt.Errorf("histories not found: %w", err)
	}
	return ViewSignupApplicationResponse{
		Application: app,
		User:        user.User,
		Assignee:    assignees,
		Histories:   histories,
	}, nil
}
