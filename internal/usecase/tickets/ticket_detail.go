// similar with view_signup_application.go
// but the response need to be create a new struct

package tickets

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/jekiapp/topic-master/internal/model/acl"
	userrepo "github.com/jekiapp/topic-master/internal/repository/user"
	dbpkg "github.com/jekiapp/topic-master/pkg/db"
	"github.com/tidwall/buntdb"
)

// Request struct for viewing ticket detail
// Only needs TicketID (which is ApplicationID)
type TicketDetailRequest struct {
	TicketID string `json:"ticket_id"`
}

type TicketDetailResponse struct {
	Ticket    ticketResponse    `json:"ticket"`
	Applicant acl.User          `json:"applicant"`
	Assignees []TicketAssignee  `json:"assignees"`
	Histories []historyResponse `json:"histories"`
	CreatedAt string            `json:"created_at"`
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

type iTicketDetailRepo interface {
	GetApplicationByID(id string) (acl.Application, error)
	GetUserByID(id string) (acl.User, error)
	ListAssignmentsByApplicationID(appID string) ([]acl.ApplicationAssignment, error)
	ListHistoriesByApplicationID(appID string) ([]acl.ApplicationHistory, error)
	GetPermissionByID(id string) (acl.Permission, error)
}

type ticketDetailRepo struct {
	db *buntdb.DB
}

func (r *ticketDetailRepo) GetApplicationByID(id string) (acl.Application, error) {
	return dbpkg.GetByID[acl.Application](r.db, id)
}

func (r *ticketDetailRepo) GetUserByID(id string) (acl.User, error) {
	return userrepo.GetUserByID(r.db, id)
}

func (r *ticketDetailRepo) ListAssignmentsByApplicationID(appID string) ([]acl.ApplicationAssignment, error) {
	return dbpkg.SelectAll[acl.ApplicationAssignment](r.db, "="+appID, acl.IdxAppAssign_ApplicationID)
}

func (r *ticketDetailRepo) ListHistoriesByApplicationID(appID string) ([]acl.ApplicationHistory, error) {
	return dbpkg.SelectAll[acl.ApplicationHistory](r.db, "="+appID, acl.IdxAppHistory_ApplicationID)
}

func (r *ticketDetailRepo) GetPermissionByID(id string) (acl.Permission, error) {
	return dbpkg.GetByID[acl.Permission](r.db, id)
}

type TicketDetailUsecase struct {
	repo iTicketDetailRepo
}

func NewTicketDetailUsecase(db *buntdb.DB) TicketDetailUsecase {
	return TicketDetailUsecase{
		repo: &ticketDetailRepo{db: db},
	}
}

func (uc TicketDetailUsecase) Handle(ctx context.Context, req map[string]string) (TicketDetailResponse, error) {
	ticketID := req["id"]
	app, err := uc.repo.GetApplicationByID(ticketID)
	if err != nil {
		return TicketDetailResponse{}, errors.New("ticket not found")
	}
	user, err := uc.repo.GetUserByID(app.UserID)
	if err != nil {
		return TicketDetailResponse{}, errors.New("user not found")
	}

	assignments, err := uc.repo.ListAssignmentsByApplicationID(ticketID)
	if err != nil {
		return TicketDetailResponse{}, fmt.Errorf("assignments not found: %w", err)
	}
	assignees := []TicketAssignee{}
	for _, assignment := range assignments {
		user, err := uc.repo.GetUserByID(assignment.ReviewerID)
		if err != nil {
			log.Println("error getting username by user id", err)
		}
		assignees = append(assignees, TicketAssignee{
			UserID:   user.ID,
			Username: user.Username,
			Name:     user.Name,
			Status:   assignment.ReviewStatus,
		})
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
		Applicant: user,
		Assignees: assignees,
		Histories: historiesResponse,
		CreatedAt: app.CreatedAt.Format(time.RFC822Z),
	}

	permissions := []permissionResponse{}
	for _, permissionID := range app.PermissionIDs {
		if strings.Contains(permissionID, "signup") {
			permissions = append(permissions, permissionResponse{
				Name:        permissionID,
				Description: "Signup application",
			})
			continue
		}

		permission, err := uc.repo.GetPermissionByID(permissionID)
		if err != nil {
			return TicketDetailResponse{}, fmt.Errorf("permission not found: %w", err)
		}
		permissions = append(permissions, permissionResponse{
			Name:        permission.Name,
			Description: permission.Description,
		})
	}
	response.Ticket.Permissions = permissions
	return response, nil
}
