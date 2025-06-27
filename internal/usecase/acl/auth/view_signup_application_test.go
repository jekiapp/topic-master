package acl

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jekiapp/topic-master/internal/model/acl"
	"github.com/jekiapp/topic-master/internal/usecase/acl/auth/mock"
	"github.com/stretchr/testify/assert"
)

func TestViewSignupApplicationUsecase_Handle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockIViewSignupApplicationRepo(ctrl)
	uc := ViewSignupApplicationUsecase{repo: mockRepo}

	tests := []struct {
		name    string
		appID   string
		setup   func()
		wantErr bool
		errMsg  string
	}{
		{
			name:  "success",
			appID: "app-1",
			setup: func() {
				app := acl.Application{ID: "app-1", UserID: "user-1"}
				userPending := acl.UserPending{User: acl.User{ID: "user-1", Username: "testuser", Name: "Test User"}}
				assignments := []acl.ApplicationAssignment{{ReviewerID: "reviewer-1", ReviewStatus: "pending"}}
				user := acl.User{ID: "reviewer-1", Username: "reviewer", Name: "Reviewer"}
				histories := []acl.ApplicationHistory{{ID: "hist-1", ApplicationID: "app-1"}}

				mockRepo.EXPECT().GetApplicationByID("app-1").Return(app, nil)
				mockRepo.EXPECT().GetUserPendingByID("user-1").Return(userPending, nil)
				mockRepo.EXPECT().ListAssignmentsByApplicationID("app-1").Return(assignments, nil)
				mockRepo.EXPECT().GetUserByID("reviewer-1").Return(user, nil)
				mockRepo.EXPECT().ListHistoriesByApplicationID("app-1").Return(histories, nil)
			},
			wantErr: false,
		},
		{
			name:  "application not found",
			appID: "notfound",
			setup: func() {
				mockRepo.EXPECT().GetApplicationByID("notfound").Return(acl.Application{}, errors.New("not found"))
			},
			wantErr: true,
			errMsg:  "application not found",
		},
		{
			name:  "user not found",
			appID: "app-2",
			setup: func() {
				app := acl.Application{ID: "app-2", UserID: "user-2"}
				mockRepo.EXPECT().GetApplicationByID("app-2").Return(app, nil)
				mockRepo.EXPECT().GetUserPendingByID("user-2").Return(acl.UserPending{}, errors.New("not found"))
			},
			wantErr: true,
			errMsg:  "user not found",
		},
		{
			name:  "assignments error",
			appID: "app-3",
			setup: func() {
				app := acl.Application{ID: "app-3", UserID: "user-3"}
				userPending := acl.UserPending{User: acl.User{ID: "user-3"}}
				mockRepo.EXPECT().GetApplicationByID("app-3").Return(app, nil)
				mockRepo.EXPECT().GetUserPendingByID("user-3").Return(userPending, nil)
				mockRepo.EXPECT().ListAssignmentsByApplicationID("app-3").Return(nil, errors.New("db error"))
			},
			wantErr: true,
			errMsg:  "assignments not found",
		},
		{
			name:  "histories error",
			appID: "app-4",
			setup: func() {
				app := acl.Application{ID: "app-4", UserID: "user-4"}
				userPending := acl.UserPending{User: acl.User{ID: "user-4"}}
				assignments := []acl.ApplicationAssignment{{ReviewerID: "reviewer-4", ReviewStatus: "pending"}}
				user := acl.User{ID: "reviewer-4", Username: "reviewer4", Name: "Reviewer Four"}
				mockRepo.EXPECT().GetApplicationByID("app-4").Return(app, nil)
				mockRepo.EXPECT().GetUserPendingByID("user-4").Return(userPending, nil)
				mockRepo.EXPECT().ListAssignmentsByApplicationID("app-4").Return(assignments, nil)
				mockRepo.EXPECT().GetUserByID("reviewer-4").Return(user, nil)
				mockRepo.EXPECT().ListHistoriesByApplicationID("app-4").Return(nil, errors.New("db error"))
			},
			wantErr: true,
			errMsg:  "histories not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			resp, err := uc.Handle(context.Background(), map[string]string{"id": tt.appID})
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.appID, resp.Application.ID)
		})
	}
}
