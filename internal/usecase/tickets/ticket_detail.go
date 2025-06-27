// similar with view_signup_application.go
// but the response need to be create a new struct

package tickets

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jekiapp/topic-master/internal/model/acl"
	entitymodel "github.com/jekiapp/topic-master/internal/model/entity"
	dbpkg "github.com/jekiapp/topic-master/pkg/db"
	"github.com/jekiapp/topic-master/pkg/util"
	"github.com/tidwall/buntdb"
)

// Request struct for viewing ticket detail
// Only needs TicketID (which is ApplicationID)
type TicketDetailRequest struct {
	TicketID string `json:"ticket_id"`
}

type TicketDetailResponse struct {
	Ticket          ticketResponse    `json:"ticket"`
	Applicant       acl.User          `json:"applicant"`
	Assignees       []TicketAssignee  `json:"assignees"`
	Histories       []historyResponse `json:"histories"`
	CreatedAt       string            `json:"created_at"`
	EligibleActions []acl.AppAction   `json:"eligible_actions"`
}

type historyResponse struct {
	Action    string `json:"action"`
	Actor     string `json:"actor"`
	Comment   string `json:"comment"`
	CreatedAt string `json:"created_at"`
}

type ticketResponse struct {
	ID          string               `json:"id"`
	Title       string               `json:"title"`
	Reason      string               `json:"reason"`
	Status      string               `json:"status"`
	Permissions []permissionResponse `json:"permissions"`
}

type permissionResponse struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type TicketAssignee struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Status   string `json:"status"`
}

func NewTicketDetailUsecase(db *buntdb.DB) TicketDetailUsecase {
	return TicketDetailUsecase{
		repo: &ticketDetailRepo{db: db},
	}
}

func (uc TicketDetailUsecase) Handle(ctx context.Context, req map[string]string) (TicketDetailResponse, error) {
	iam := util.GetUserInfo(ctx)

	ticketID := req["id"]
	app, err := uc.repo.GetApplicationByID(ticketID)
	if err != nil {
		return TicketDetailResponse{}, fmt.Errorf("ticket %s not found", ticketID)
	}

	// if the permission is not signup (check the first permissionIDs, then get user by id)
	var applicant acl.User
	if len(app.PermissionIDs) > 0 && strings.HasPrefix(app.PermissionIDs[0], "signup:") {
		userPending, err := uc.repo.GetUserPendingByID(app.UserID)
		if err != nil {
			return TicketDetailResponse{}, fmt.Errorf("user %s not found", app.UserID)
		}
		applicant = userPending.User
	} else {
		user, err := uc.repo.GetUserByID(app.UserID)
		if err != nil {
			return TicketDetailResponse{}, fmt.Errorf("user %s not found", app.UserID)
		}
		applicant = user
	}

	assignments, err := uc.repo.ListAssignmentsByApplicationID(ticketID)
	if err != nil {
		return TicketDetailResponse{}, fmt.Errorf("assignments not found: %w", err)
	}

	eligible := false
	assignees := []TicketAssignee{}
	for _, assignment := range assignments {
		user, err := uc.repo.GetUserByID(assignment.ReviewerID)
		if err != nil {
			log.Println("error getting username by user id", err)
		}
		if user.ID == iam.ID {
			eligible = true
		}
		assignees = append(assignees, TicketAssignee{
			UserID:   user.ID,
			Username: user.Username,
			Name:     user.Name,
			Status:   assignment.ReviewStatus,
		})
	}

	if app.Status == acl.StatusCompleted {
		eligible = false
	}

	eligibleActions := []acl.AppAction{}
	if eligible {
		eligibleActions = append(eligibleActions, acl.AppActionApprove)
		eligibleActions = append(eligibleActions, acl.AppActionReject)
	}

	histories, err := uc.repo.ListHistoriesByApplicationID(ticketID)
	if err != nil {
		return TicketDetailResponse{}, fmt.Errorf("histories not found: %w", err)
	}

	historiesResponse := []historyResponse{}
	for _, history := range histories {
		actor, err := uc.repo.GetUserByID(history.ActorID)
		if err != nil {
			log.Println("error getting username by user id", err)
			// fallback to pending user
			pendingActor, err := uc.repo.GetUserPendingByID(history.ActorID)
			if err != nil {
				log.Println("error getting username by user id", err)
			} else {
				actor = pendingActor.User
			}
		}
		historiesResponse = append(historiesResponse, historyResponse{
			Action:    history.Action,
			Actor:     actor.Name,
			Comment:   history.Comment,
			CreatedAt: history.CreatedAt.Format(time.RFC822Z),
		})
	}

	response := TicketDetailResponse{
		Ticket: ticketResponse{
			ID:     app.ID,
			Title:  app.Title,
			Reason: app.Reason,
			Status: app.Status,
		},
		Applicant:       applicant,
		Assignees:       assignees,
		Histories:       historiesResponse,
		CreatedAt:       app.CreatedAt.Format(time.RFC822Z),
		EligibleActions: eligibleActions,
	}

	permissions := []permissionResponse{}
	for _, permissionID := range app.PermissionIDs {
		if strings.HasPrefix(permissionID, "signup") {
			permissions = append(permissions, permissionResponse{
				Name:        permissionID,
				Description: "Signup application",
			})
			continue
		}

		if strings.HasPrefix(permissionID, "claim") {
			entityID := strings.TrimPrefix(permissionID, "claim:")
			entity, err := uc.repo.GetEntityByID(entityID)
			if err != nil {
				return TicketDetailResponse{}, fmt.Errorf("entity not found: %w", err)
			}
			groupName := app.MetaData[entityID+":group_name"]
			permissions = append(permissions, permissionResponse{
				Name:        "claim:" + entity.Name,
				Description: groupName,
			})
			continue
		}
	}
	response.Ticket.Permissions = permissions
	return response, nil
}

type iTicketDetailRepo interface {
	GetApplicationByID(id string) (acl.Application, error)
	GetUserByID(id string) (acl.User, error)
	GetUserPendingByID(id string) (acl.UserPending, error)
	ListAssignmentsByApplicationID(appID string) ([]acl.ApplicationAssignment, error)
	ListHistoriesByApplicationID(appID string) ([]acl.ApplicationHistory, error)
	GetPermissionByID(id string) (acl.PermissionMap, error)
	GetEntityByID(id string) (entitymodel.Entity, error)
}

type ticketDetailRepo struct {
	db *buntdb.DB
}

func (r *ticketDetailRepo) GetApplicationByID(id string) (acl.Application, error) {
	return dbpkg.GetByID[acl.Application](r.db, id)
}

func (r *ticketDetailRepo) GetUserByID(id string) (acl.User, error) {
	return dbpkg.GetByID[acl.User](r.db, id)
}

func (r *ticketDetailRepo) GetUserPendingByID(id string) (acl.UserPending, error) {
	return dbpkg.GetByID[acl.UserPending](r.db, id)
}

func (r *ticketDetailRepo) ListAssignmentsByApplicationID(appID string) ([]acl.ApplicationAssignment, error) {
	return dbpkg.SelectAll[acl.ApplicationAssignment](r.db, "="+appID, acl.IdxAppAssign_ApplicationID)
}

func (r *ticketDetailRepo) ListHistoriesByApplicationID(appID string) ([]acl.ApplicationHistory, error) {
	appID = fmt.Sprintf("%s:%d", appID, time.Now().Unix())
	return dbpkg.SelectAll[acl.ApplicationHistory](r.db, "-<="+appID, acl.IdxAppHistory_ApplicationID)
}

func (r *ticketDetailRepo) GetPermissionByID(id string) (acl.PermissionMap, error) {
	return dbpkg.GetByID[acl.PermissionMap](r.db, id)
}

func (r *ticketDetailRepo) GetEntityByID(id string) (entitymodel.Entity, error) {
	return dbpkg.GetByID[entitymodel.Entity](r.db, id)
}

type TicketDetailUsecase struct {
	repo iTicketDetailRepo
}
