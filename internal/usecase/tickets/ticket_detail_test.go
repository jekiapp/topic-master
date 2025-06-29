package tickets

import (
	"context"
	"errors"
	"testing"

	"github.com/jekiapp/topic-master/internal/model/acl"
	"github.com/jekiapp/topic-master/internal/usecase/tickets/mock"
	"github.com/jekiapp/topic-master/pkg/util"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func mockContextWithUser(ctx context.Context, user *acl.User) context.Context {
	return util.MockContextWithUser(ctx, user)
}

func TestTicketDetailUsecase_Handle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	bob := &acl.User{ID: "bob-id", Username: "bob"}
	alice := &acl.User{ID: "alice-id", Name: "Alice", Username: "alice"}
	carol := &acl.User{ID: "carol-id", Name: "Carol", Username: "carol"}
	app := acl.Application{ID: "app-bob-1", UserID: "alice-id", Title: "Alice's Application", Status: "open", PermissionIDs: []string{"perm1"}}
	assignment := acl.ApplicationAssignment{ApplicationID: "app-bob-1", ReviewerID: "carol-id", ReviewStatus: "pending"}
	history := acl.ApplicationHistory{Action: "create", ActorID: "alice-id", Comment: "init", CreatedAt: app.CreatedAt}

	tests := []struct {
		name      string
		user      *acl.User
		req       map[string]string
		setupMock func(m *mock.MockiTicketDetailRepo)
		wantErr   bool
		wantResp  TicketDetailResponse
	}{
		{
			name: "ticket not found",
			user: bob,
			req:  map[string]string{"id": "notfound-id"},
			setupMock: func(m *mock.MockiTicketDetailRepo) {
				m.EXPECT().GetApplicationByID("notfound-id").Return(acl.Application{}, errors.New("not found"))
			},
			wantErr: true,
		},
		{
			name: "applicant not found",
			user: bob,
			req:  map[string]string{"id": "app-bob-1"},
			setupMock: func(m *mock.MockiTicketDetailRepo) {
				m.EXPECT().GetApplicationByID("app-bob-1").Return(app, nil)
				m.EXPECT().GetUserByID("alice-id").Return(acl.User{}, errors.New("not found"))
			},
			wantErr: true,
		},
		{
			name: "assignments error",
			user: bob,
			req:  map[string]string{"id": "app-bob-1"},
			setupMock: func(m *mock.MockiTicketDetailRepo) {
				m.EXPECT().GetApplicationByID("app-bob-1").Return(app, nil)
				m.EXPECT().GetUserByID("alice-id").Return(*alice, nil)
				m.EXPECT().ListAssignmentsByApplicationID("app-bob-1").Return(nil, errors.New("assignments error"))
			},
			wantErr: true,
		},
		{
			name: "histories error",
			user: bob,
			req:  map[string]string{"id": "app-bob-1"},
			setupMock: func(m *mock.MockiTicketDetailRepo) {
				m.EXPECT().GetApplicationByID("app-bob-1").Return(app, nil)
				m.EXPECT().GetUserByID("alice-id").Return(*alice, nil)
				m.EXPECT().ListAssignmentsByApplicationID("app-bob-1").Return([]acl.ApplicationAssignment{assignment}, nil)
				m.EXPECT().GetUserByID("carol-id").Return(*carol, nil)
				m.EXPECT().ListHistoriesByApplicationID("app-bob-1").Return(nil, errors.New("histories error"))
			},
			wantErr: true,
		},
		{
			name: "success",
			user: bob,
			req:  map[string]string{"id": "app-bob-1"},
			setupMock: func(m *mock.MockiTicketDetailRepo) {
				m.EXPECT().GetApplicationByID("app-bob-1").Return(app, nil)
				m.EXPECT().GetUserByID("alice-id").Return(*alice, nil)
				m.EXPECT().ListAssignmentsByApplicationID("app-bob-1").Return([]acl.ApplicationAssignment{assignment}, nil)
				m.EXPECT().GetUserByID("carol-id").Return(*carol, nil)
				m.EXPECT().ListHistoriesByApplicationID("app-bob-1").Return([]acl.ApplicationHistory{history}, nil)
			},
			wantResp: TicketDetailResponse{
				Ticket:          ticketResponse{ID: "app-bob-1", Title: "Alice's Application", Reason: "", Status: "open"},
				Applicant:       *alice,
				Assignees:       []TicketAssignee{{UserID: "carol-id", Username: "carol", Name: "Carol", Status: "pending"}},
				Histories:       []historyResponse{{Action: "create", Actor: "Alice", Comment: "init", CreatedAt: ""}},
				CreatedAt:       "",
				EligibleActions: []acl.AppAction{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mock.NewMockiTicketDetailRepo(ctrl)
			tt.setupMock(mockRepo)
			uc := TicketDetailUsecase{repo: mockRepo}
			ctx := context.Background()
			if tt.user != nil {
				ctx = util.MockContextWithUser(ctx, tt.user)
			}
			resp, err := uc.Handle(ctx, tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantResp.Ticket.ID, resp.Ticket.ID)
				assert.Equal(t, tt.wantResp.Applicant, resp.Applicant)
				assert.Equal(t, tt.wantResp.Assignees, resp.Assignees)
			}
		})
	}
}
