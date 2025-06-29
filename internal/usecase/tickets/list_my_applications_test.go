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

func TestListMyApplicationsUsecase_Handle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	bob := &acl.User{ID: "bob-id", Username: "bob"}
	apps := []acl.Application{{ID: "app-bob-1", UserID: "bob-id", Title: "Bob's Application"}}
	assignments := []acl.ApplicationAssignment{{ApplicationID: "app-bob-1", ReviewerID: "alice-id"}}
	assignee := acl.User{ID: "alice-id", Username: "alice"}

	tests := []struct {
		name      string
		user      *acl.User
		req       map[string]string
		setupMock func(m *mock.MockiMyApplicationRepo)
		wantErr   bool
		wantResp  ListMyApplicationsResponse
	}{
		{
			name:      "unauthorized user",
			user:      nil,
			req:       map[string]string{},
			setupMock: func(m *mock.MockiMyApplicationRepo) {},
			wantErr:   true,
		},
		{
			name: "repo error",
			user: bob,
			req:  map[string]string{"page": "1", "limit": "10"},
			setupMock: func(m *mock.MockiMyApplicationRepo) {
				m.EXPECT().ListApplicationsByUserIDDescPaginated("bob-id", 1, 10).Return(nil, false, assert.AnError)
			},
			wantErr: true,
		},
		{
			name: "empty result",
			user: bob,
			req:  map[string]string{},
			setupMock: func(m *mock.MockiMyApplicationRepo) {
				m.EXPECT().ListApplicationsByUserIDDescPaginated("bob-id", 1, 10).Return([]acl.Application{}, false, nil)
			},
			wantResp: ListMyApplicationsResponse{Applications: nil, HasNext: false},
		},
		{
			name: "success with assignments",
			user: bob,
			req:  map[string]string{},
			setupMock: func(m *mock.MockiMyApplicationRepo) {
				m.EXPECT().ListApplicationsByUserIDDescPaginated("bob-id", 1, 10).Return(apps, false, nil)
				m.EXPECT().GetUserByID("bob-id").Return(*bob, nil)
				m.EXPECT().ListAssignmentsByApplicationID("app-bob-1").Return(assignments, nil)
				m.EXPECT().GetUserByID("alice-id").Return(assignee, nil)
			},
			wantResp: ListMyApplicationsResponse{
				Applications: []myApplicationResponse{{Application: apps[0], ApplicantName: "bob", AssigneeNames: "alice"}},
				HasNext:      false,
			},
		},
		{
			name: "assignment error",
			user: bob,
			req:  map[string]string{},
			setupMock: func(m *mock.MockiMyApplicationRepo) {
				m.EXPECT().ListApplicationsByUserIDDescPaginated("bob-id", 1, 10).Return(apps, false, nil)
				m.EXPECT().GetUserByID("bob-id").Return(*bob, nil)
				m.EXPECT().ListAssignmentsByApplicationID("app-bob-1").Return(nil, assert.AnError)
			},
			wantResp: ListMyApplicationsResponse{
				Applications: []myApplicationResponse{{Application: apps[0], ApplicantName: "bob", AssigneeNames: ""}},
				HasNext:      false,
			},
		},
		{
			name: "applicant not found",
			user: bob,
			req:  map[string]string{},
			setupMock: func(m *mock.MockiMyApplicationRepo) {
				m.EXPECT().ListApplicationsByUserIDDescPaginated("bob-id", 1, 10).Return(apps, false, nil)
				m.EXPECT().GetUserByID("bob-id").Return(acl.User{}, assert.AnError)
				m.EXPECT().ListAssignmentsByApplicationID("app-bob-1").Return(assignments, nil)
				m.EXPECT().GetUserByID("alice-id").Return(assignee, nil)
			},
			wantResp: ListMyApplicationsResponse{
				Applications: []myApplicationResponse{{Application: apps[0], ApplicantName: "", AssigneeNames: "alice"}},
				HasNext:      false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := mock.NewMockiMyApplicationRepo(ctrl)
			tt.setupMock(mockRepo)
			uc := ListMyApplicationsUsecase{repo: mockRepo}
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
