package tickets

import (
	"context"
	"testing"

	"github.com/jekiapp/topic-master/internal/model/acl"
	"github.com/jekiapp/topic-master/internal/usecase/tickets/mock"
	"github.com/jekiapp/topic-master/pkg/util"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestListMyAssignmentUsecase_Handle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	alice := &acl.User{ID: "alice-id", Username: "alice"}
	assignments := []acl.ApplicationAssignment{{ApplicationID: "app-alice-1", ReviewerID: "alice-id"}}
	app := acl.Application{ID: "app-alice-1", UserID: "bob-id", Title: "Bob's Application"}
	applicant := acl.User{ID: "bob-id", Username: "bob"}

	tests := []struct {
		name      string
		user      *acl.User
		req       map[string]string
		setupMock func(m *mock.MockiAssignmentRepo)
		wantErr   bool
		wantResp  ListMyAssignmentResponse
	}{
		{
			name:      "unauthorized user",
			user:      nil,
			req:       map[string]string{},
			setupMock: func(m *mock.MockiAssignmentRepo) {},
			wantErr:   true,
		},
		{
			name: "repo error",
			user: alice,
			req:  map[string]string{"page": "1", "limit": "20"},
			setupMock: func(m *mock.MockiAssignmentRepo) {
				m.EXPECT().ListAssignmentsByReviewerIDPaginated("alice-id", 1, 20).Return(nil, false, assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "empty result",
			user: alice,
			req:  map[string]string{},
			setupMock: func(m *mock.MockiAssignmentRepo) {
				m.EXPECT().ListAssignmentsByReviewerIDPaginated("alice-id", 1, 20).Return([]acl.ApplicationAssignment{}, false, nil)
			},
			wantResp: ListMyAssignmentResponse{Applications: nil, HasNext: false},
		},
		{
			name: "success with applicant",
			user: alice,
			req:  map[string]string{},
			setupMock: func(m *mock.MockiAssignmentRepo) {
				m.EXPECT().ListAssignmentsByReviewerIDPaginated("alice-id", 1, 20).Return(assignments, false, nil)
				m.EXPECT().GetApplicationByID("app-alice-1").Return(app, nil)
				m.EXPECT().GetUserByID("bob-id").Return(applicant, nil)
			},
			wantResp: ListMyAssignmentResponse{
				Applications: []applicationResponse{{Application: app, ApplicantName: "bob"}},
				HasNext:      false,
			},
		},
		{
			name: "app not found",
			user: alice,
			req:  map[string]string{},
			setupMock: func(m *mock.MockiAssignmentRepo) {
				m.EXPECT().ListAssignmentsByReviewerIDPaginated("alice-id", 1, 20).Return(assignments, false, nil)
				m.EXPECT().GetApplicationByID("app-alice-1").Return(acl.Application{}, assert.AnError)
			},
			wantResp: ListMyAssignmentResponse{Applications: nil, HasNext: false},
		},
		{
			name: "applicant not found",
			user: alice,
			req:  map[string]string{},
			setupMock: func(m *mock.MockiAssignmentRepo) {
				m.EXPECT().ListAssignmentsByReviewerIDPaginated("alice-id", 1, 20).Return(assignments, false, nil)
				m.EXPECT().GetApplicationByID("app-alice-1").Return(app, nil)
				m.EXPECT().GetUserByID("bob-id").Return(acl.User{}, assert.AnError)
			},
			wantResp: ListMyAssignmentResponse{
				Applications: []applicationResponse{{Application: app, ApplicantName: ""}},
				HasNext:      false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mock.NewMockiAssignmentRepo(ctrl)
			tt.setupMock(mockRepo)
			uc := ListMyAssignmentUsecase{repo: mockRepo}
			ctx := context.Background()
			if tt.user != nil {
				ctx = util.MockContextWithUser(ctx, tt.user)
			}
			resp, err := uc.Handle(ctx, tt.req)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantResp.HasNext, resp.HasNext)
				assert.Equal(t, tt.wantResp.Applications, resp.Applications)
			}
		})
	}
}
