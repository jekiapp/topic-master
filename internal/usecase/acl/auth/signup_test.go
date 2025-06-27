package acl

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/jekiapp/topic-master/internal/model/acl"
	"github.com/jekiapp/topic-master/internal/usecase/acl/auth/mock"
	"github.com/stretchr/testify/assert"
	"github.com/tidwall/buntdb"
)

func TestSignupUsecase_Handle(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mock.NewMockISignupRepo(ctrl)
	uc := SignupUsecase{repo: mockRepo}

	tests := []struct {
		name    string
		req     SignupRequest
		setup   func()
		wantErr bool
		errMsg  string
	}{
		{
			name: "success",
			req: SignupRequest{
				Username:        "testuser",
				Name:            "Test User",
				Password:        "password123",
				ConfirmPassword: "password123",
				GroupID:         "group1",
				GroupName:       "Group One",
				GroupRole:       "member",
				Reason:          "Test signup",
			},
			setup: func() {
				rootGroup := acl.Group{
					ID:   "root-group",
					Name: acl.GroupRoot,
				}
				mockRepo.EXPECT().GetGroupByName(acl.GroupRoot).Return(rootGroup, nil)

				rootMembers := []acl.UserGroup{
					{
						ID:      "ug1",
						UserID:  "root1",
						GroupID: "root-group",
					},
				}
				mockRepo.EXPECT().ListUserGroupsByGroupID(rootGroup.ID, 0).Return(rootMembers, nil)

				mockRepo.EXPECT().GetAdminUserIDsByGroupID("group1").Return([]string{"admin1"}, nil)

				mockRepo.EXPECT().GetUserByID("root1").Return(acl.User{Status: acl.StatusUserActive}, nil)
				mockRepo.EXPECT().GetUserByID("admin1").Return(acl.User{Status: acl.StatusUserActive}, nil)

				// Verify application creation
				var appID string
				mockRepo.EXPECT().CreateApplication(gomock.Any()).DoAndReturn(func(app acl.Application) error {
					assert.NotEmpty(t, app.ID)
					appID = app.ID
					assert.Contains(t, app.Title, "Signup request by Test User (testuser)")
					assert.Equal(t, acl.ApplicationType_Signup, app.Type)
					assert.NotEmpty(t, app.UserID)
					assert.Equal(t, []string{acl.Permission_Signup_User.Name}, app.PermissionIDs)
					assert.Equal(t, "Request to become member of group Group One", app.Reason)
					assert.Equal(t, acl.StatusWaitingForApproval, app.Status)
					assert.False(t, app.CreatedAt.IsZero())
					assert.False(t, app.UpdatedAt.IsZero())
					return nil
				})

				// Verify application assignments
				mockRepo.EXPECT().CreateApplicationAssignment(gomock.Any()).DoAndReturn(func(assignment acl.ApplicationAssignment) error {
					assert.NotEmpty(t, assignment.ID)
					assert.Equal(t, appID, assignment.ApplicationID)
					assert.Equal(t, "admin1", assignment.ReviewerID)
					assert.Equal(t, acl.ActionWaitingForApproval, assignment.ReviewStatus)
					assert.False(t, assignment.CreatedAt.IsZero())
					assert.False(t, assignment.UpdatedAt.IsZero())
					return nil
				})

				mockRepo.EXPECT().CreateApplicationAssignment(gomock.Any()).DoAndReturn(func(assignment acl.ApplicationAssignment) error {
					assert.NotEmpty(t, assignment.ID)
					assert.Equal(t, appID, assignment.ApplicationID)
					assert.Equal(t, "root1", assignment.ReviewerID)
					assert.Equal(t, acl.ActionWaitingForApproval, assignment.ReviewStatus)
					assert.False(t, assignment.CreatedAt.IsZero())
					assert.False(t, assignment.UpdatedAt.IsZero())
					return nil
				})

				// Verify user pending creation
				var userID string
				mockRepo.EXPECT().CreateUserPending(gomock.Any()).DoAndReturn(func(user acl.UserPending) error {
					assert.NotEmpty(t, user.ID)
					userID = user.ID
					assert.Equal(t, "testuser", user.Username)
					assert.Equal(t, "Test User", user.Name)
					// Verify password is hashed
					hash := sha256.Sum256([]byte("password123"))
					expectedPassword := hex.EncodeToString(hash[:])
					assert.Equal(t, expectedPassword, user.Password)
					assert.Equal(t, acl.StatusUserInApproval, user.Status)
					assert.Equal(t, []acl.GroupRole{{
						GroupID:   "group1",
						GroupName: "Group One",
						Role:      "member",
					}}, user.Groups)
					assert.False(t, user.CreatedAt.IsZero())
					assert.False(t, user.UpdatedAt.IsZero())
					return nil
				})

				// Verify user group creation
				mockRepo.EXPECT().CreateUserGroup(gomock.Any()).DoAndReturn(func(userGroup acl.UserGroup) error {
					assert.NotEmpty(t, userGroup.ID)
					assert.Equal(t, userID, userGroup.UserID)
					assert.Equal(t, "group1", userGroup.GroupID)
					assert.Equal(t, "member", userGroup.Role)
					assert.False(t, userGroup.CreatedAt.IsZero())
					assert.False(t, userGroup.UpdatedAt.IsZero())
					return nil
				})

				// Verify application history creation
				mockRepo.EXPECT().CreateApplicationHistory(gomock.Any()).DoAndReturn(func(history acl.ApplicationHistory) error {
					assert.NotEmpty(t, history.ID)
					assert.Equal(t, appID, history.ApplicationID)
					assert.Equal(t, "Create ticket", history.Action)
					assert.Equal(t, userID, history.ActorID)
					assert.Equal(t, "Initial signup by Test User", history.Comment)
					assert.False(t, history.CreatedAt.IsZero())
					assert.False(t, history.UpdatedAt.IsZero())
					return nil
				})
			},
			wantErr: false,
		},
		{
			name: "validation error - missing username",
			req: SignupRequest{
				Name:            "Test User",
				Password:        "password123",
				ConfirmPassword: "password123",
				GroupID:         "group1",
				GroupRole:       "member",
			},
			setup:   func() {},
			wantErr: true,
			errMsg:  "missing username",
		},
		{
			name: "validation error - password mismatch",
			req: SignupRequest{
				Username:        "testuser",
				Name:            "Test User",
				Password:        "password123",
				ConfirmPassword: "different",
				GroupID:         "group1",
				GroupRole:       "member",
			},
			setup:   func() {},
			wantErr: true,
			errMsg:  "password and confirm_password do not match",
		},
		{
			name: "error - root group not found",
			req: SignupRequest{
				Username:        "testuser",
				Name:            "Test User",
				Password:        "password123",
				ConfirmPassword: "password123",
				GroupID:         "group1",
				GroupName:       "Group One",
				GroupRole:       "member",
			},
			setup: func() {
				mockRepo.EXPECT().CreateApplication(gomock.Any()).DoAndReturn(func(app acl.Application) error {
					assert.NotEmpty(t, app.ID)
					assert.Contains(t, app.Title, "Signup request by Test User (testuser)")
					return nil
				})
				mockRepo.EXPECT().GetGroupByName(acl.GroupRoot).Return(acl.Group{}, buntdb.ErrNotFound)
			},
			wantErr: true,
			errMsg:  "root group not found",
		},
		{
			name: "error - inactive reviewer",
			req: SignupRequest{
				Username:        "testuser",
				Name:            "Test User",
				Password:        "password123",
				ConfirmPassword: "password123",
				GroupID:         "group1",
				GroupName:       "Group One",
				GroupRole:       "member",
			},
			setup: func() {
				mockRepo.EXPECT().CreateApplication(gomock.Any()).DoAndReturn(func(app acl.Application) error {
					assert.NotEmpty(t, app.ID)
					assert.Contains(t, app.Title, "Signup request by Test User (testuser)")
					return nil
				})

				rootGroup := acl.Group{
					ID:   "root-group",
					Name: acl.GroupRoot,
				}
				mockRepo.EXPECT().GetGroupByName(acl.GroupRoot).Return(rootGroup, nil)

				rootMembers := []acl.UserGroup{
					{
						ID:      "ug1",
						UserID:  "root1",
						GroupID: "root-group",
					},
				}
				mockRepo.EXPECT().ListUserGroupsByGroupID(rootGroup.ID, 0).Return(rootMembers, nil)

				mockRepo.EXPECT().GetAdminUserIDsByGroupID("group1").Return([]string{"admin1"}, nil)

				// Both reviewers are inactive
				mockRepo.EXPECT().GetUserByID("root1").Return(acl.User{Status: acl.StatusUserInactive}, nil)
				mockRepo.EXPECT().GetUserByID("admin1").Return(acl.User{Status: acl.StatusUserInactive}, nil)
			},
			wantErr: true,
			errMsg:  "no active reviewers found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setup()
			resp, err := uc.Handle(context.Background(), tt.req)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				return
			}
			assert.NoError(t, err)
			assert.NotEmpty(t, resp.ApplicationID)
		})
	}
}
