package tickets

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jekiapp/topic-master/internal/model/acl"
	dbpkg "github.com/jekiapp/topic-master/pkg/db"
	"github.com/jekiapp/topic-master/pkg/util"
	"github.com/tidwall/buntdb"
)

type ListMyApplicationsRequest struct{}

type ListMyApplicationsResponse struct {
	Applications []myApplicationResponse `json:"applications"`
	HasNext      bool                    `json:"has_next"`
}

type iMyApplicationRepo interface {
	ListApplicationsByUserIDDescPaginated(userID string, page, limit int) ([]acl.Application, bool, error)
	GetUserByID(userID string) (acl.User, error)
	ListAssignmentsByApplicationID(appID string) ([]acl.ApplicationAssignment, error)
}

type myApplicationResponse struct {
	acl.Application
	ApplicantName string `json:"applicant_name"`
	AssigneeNames string `json:"assignee_names"`
}

type applicationRepoImpl struct {
	db *buntdb.DB
}

func (r *applicationRepoImpl) ListApplicationsByUserIDDescPaginated(userID string, page, limit int) ([]acl.Application, bool, error) {
	pivot := fmt.Sprintf("%s:%d", userID, time.Now().Unix())
	pagination := &dbpkg.Pagination{Page: page, Limit: limit}
	apps, err := dbpkg.SelectPaginated[acl.Application](r.db, "-<="+pivot, acl.IdxApplication_UserID_CreatedAt, pagination)
	if err != nil && err != dbpkg.ErrNotFound {
		return nil, false, err
	}
	return apps, pagination.HasNext, nil
}

func (r *applicationRepoImpl) GetUserByID(userID string) (acl.User, error) {
	return dbpkg.GetByID[acl.User](r.db, userID)
}

func (r *applicationRepoImpl) ListAssignmentsByApplicationID(appID string) ([]acl.ApplicationAssignment, error) {
	return dbpkg.SelectAll[acl.ApplicationAssignment](r.db, "="+appID, acl.IdxAppAssign_ApplicationID)
}

type ListMyApplicationsUsecase struct {
	repo iMyApplicationRepo
}

func NewListMyApplicationsUsecase(db *buntdb.DB) ListMyApplicationsUsecase {
	return ListMyApplicationsUsecase{
		repo: &applicationRepoImpl{db: db},
	}
}

func (uc ListMyApplicationsUsecase) Handle(ctx context.Context, req map[string]string) (ListMyApplicationsResponse, error) {
	user := util.GetUserInfo(ctx)
	if user == nil {
		return ListMyApplicationsResponse{}, fmt.Errorf("unauthorized: user info not found")
	}
	// parse page and limit from req
	page, limit := 1, 10
	if v, ok := req["page"]; ok {
		fmt.Sscanf(v, "%d", &page)
	}
	if v, ok := req["limit"]; ok {
		fmt.Sscanf(v, "%d", &limit)
	}
	apps, hasNext, err := uc.repo.ListApplicationsByUserIDDescPaginated(user.ID, page, limit)
	if err != nil && err != dbpkg.ErrNotFound {
		return ListMyApplicationsResponse{}, err
	}
	appsResp := []myApplicationResponse{}
	for _, app := range apps {
		applicant, err := uc.repo.GetUserByID(app.UserID)
		name := ""
		if err == nil {
			name = applicant.Username
		}
		// Fetch assignees (reviewers)
		assignments, err := uc.repo.ListAssignmentsByApplicationID(app.ID)
		if err != nil {
			log.Println("error listing assignments by application id", err)
		}

		assigneeNames := []string{}
		for _, assign := range assignments {
			reviewer, err := uc.repo.GetUserByID(assign.ReviewerID)
			if err == nil {
				assigneeNames = append(assigneeNames, reviewer.Username)
			}
		}
		appsResp = append(appsResp, myApplicationResponse{
			Application:   app,
			ApplicantName: name,
			AssigneeNames: joinNames(assigneeNames),
		})
	}
	return ListMyApplicationsResponse{Applications: appsResp, HasNext: hasNext}, nil
}

// joinNames joins names with comma, returns empty string if none
func joinNames(names []string) string {
	if len(names) == 0 {
		return ""
	}
	return strings.Join(names, ", ")
}
