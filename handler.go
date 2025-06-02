package main

import (
	"net/http"

	"github.com/jekiapp/topic-master/internal/config"
	acl "github.com/jekiapp/topic-master/internal/usecase/acl"
	aclAuth "github.com/jekiapp/topic-master/internal/usecase/acl/auth"
	aclGroup "github.com/jekiapp/topic-master/internal/usecase/acl/group"
	aclUser "github.com/jekiapp/topic-master/internal/usecase/acl/user"
	aclUserGroup "github.com/jekiapp/topic-master/internal/usecase/acl/usergroup"
	"github.com/jekiapp/topic-master/internal/usecase/tickets"
	topicUC "github.com/jekiapp/topic-master/internal/usecase/topic"
	webUC "github.com/jekiapp/topic-master/internal/usecase/web"
	handlerPkg "github.com/jekiapp/topic-master/pkg/handler"
	"github.com/tidwall/buntdb"
)

type Handler struct {
	config                  *config.Config
	createUserUC            aclUser.CreateUserUsecase
	updateUserUC            aclUser.UpdateUserUsecase
	loginUC                 aclAuth.LoginUsecase
	logoutUC                aclAuth.LogoutUsecase
	assignUserToGroupUC     aclUserGroup.AssignUserToGroupUsecase
	deleteUserUC            aclUser.DeleteUserUsecase
	createGroupUC           aclGroup.CreateGroupUsecase
	createPermissionUC      acl.CreatePermissionUsecase
	changePasswordUC        aclUser.ChangePasswordUsecase
	syncTopicsUC            topicUC.SyncTopicsUsecase
	webUC                   *webUC.WebUsecase
	getGroupListUC          aclGroup.GetGroupListUsecase
	getGroupListSimpleUC    aclGroup.GetGroupListSimpleUsecase
	getUserListUC           aclUser.GetUserListUsecase
	listTopicsUC            topicUC.ListTopicsUsecase
	updateGroupByIDUC       aclGroup.UpdateGroupByIDUsecase
	deleteGroupUC           aclGroup.DeleteGroupUsecase
	resetPasswordUC         aclAuth.ResetPasswordUsecase
	signupUC                aclAuth.SignupUsecase
	viewSignupApplicationUC aclAuth.ViewSignupApplicationUsecase
	listMyAssignmentUC      tickets.ListMyAssignmentUsecase
}

func initHandler(db *buntdb.DB, cfg *config.Config) Handler {
	webUsecase := webUC.NewWebUsecase()

	return Handler{
		config:                  cfg,
		createUserUC:            aclUser.NewCreateUserUsecase(db),
		updateUserUC:            aclUser.NewUpdateUserUsecase(db),
		loginUC:                 aclAuth.NewLoginUsecase(db, cfg),
		logoutUC:                aclAuth.NewLogoutUsecase(),
		assignUserToGroupUC:     aclUserGroup.NewAssignUserToGroupUsecase(db),
		deleteUserUC:            aclUser.NewDeleteUserUsecase(db),
		createGroupUC:           aclGroup.NewCreateGroupUsecase(db),
		createPermissionUC:      acl.NewCreatePermissionUsecase(db),
		changePasswordUC:        aclUser.NewChangePasswordUsecase(db),
		syncTopicsUC:            topicUC.NewSyncTopicsUsecase(db),
		webUC:                   webUsecase,
		getGroupListUC:          aclGroup.NewGetGroupListUsecase(db),
		getGroupListSimpleUC:    aclGroup.NewGetGroupListSimpleUsecase(db),
		getUserListUC:           aclUser.NewGetUserListUsecase(db),
		listTopicsUC:            topicUC.NewListTopicsUsecase(db),
		updateGroupByIDUC:       aclGroup.NewUpdateGroupByIDUsecase(db),
		deleteGroupUC:           aclGroup.NewDeleteGroupUsecase(db),
		resetPasswordUC:         aclAuth.NewResetPasswordUsecase(db),
		signupUC:                aclAuth.NewSignupUsecase(db),
		viewSignupApplicationUC: aclAuth.NewViewSignupApplicationUsecase(db),
		listMyAssignmentUC:      tickets.NewListMyAssignmentUsecase(db),
	}
}

func (h Handler) routes(mux *http.ServeMux) {
	authMiddleware := handlerPkg.InitJWTMiddleware(string(h.config.SecretKey))

	mux.HandleFunc("/api/login", h.loginUC.Handle)
	mux.HandleFunc("/api/logout", h.logoutUC.Handle)

	mux.HandleFunc("/api/user/list", handlerPkg.HandleGenericPost(h.getUserListUC.Handle))
	mux.HandleFunc("/api/user/create", authMiddleware(handlerPkg.HandleGenericPost(h.createUserUC.Handle)))
	mux.HandleFunc("/api/user/update", authMiddleware(handlerPkg.HandleGenericPost(h.updateUserUC.Handle)))
	mux.HandleFunc("/api/user/assign-to-group", authMiddleware(handlerPkg.HandleGenericPost(h.assignUserToGroupUC.Handle)))
	mux.HandleFunc("/api/user/delete", authMiddleware(handlerPkg.HandleGenericPost(h.deleteUserUC.Handle)))

	mux.HandleFunc("/api/create-permission", handlerPkg.HandleGenericPost(h.createPermissionUC.Handle))
	mux.HandleFunc("/api/change-password", handlerPkg.HandleGenericPost(h.changePasswordUC.Handle))
	mux.HandleFunc("/api/sync-topics", handlerPkg.QueryHandler(h.syncTopicsUC.HandleQuery))

	mux.HandleFunc("/api/group/create", authMiddleware(handlerPkg.HandleGenericPost(h.createGroupUC.Handle)))
	mux.HandleFunc("/api/group/list", authMiddleware(handlerPkg.HandleGenericPost(h.getGroupListUC.Handle)))
	mux.HandleFunc("/api/group/list-simple", handlerPkg.HandleGenericPost(h.getGroupListSimpleUC.Handle))
	mux.HandleFunc("/api/group/update-group-by-id", authMiddleware(handlerPkg.HandleGenericPost(h.updateGroupByIDUC.Handle)))
	mux.HandleFunc("/api/group/delete-group", authMiddleware(handlerPkg.HandleGenericPost(h.deleteGroupUC.Handle)))

	mux.HandleFunc("/api/topic/list-topics", authMiddleware(handlerPkg.QueryHandler(h.listTopicsUC.HandleQuery)))

	mux.HandleFunc("/api/reset-password", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			handlerPkg.QueryHandler(h.resetPasswordUC.HandleGet)(w, r)
			return
		}
		if r.Method == http.MethodPost {
			handlerPkg.HandleGenericPost(h.resetPasswordUC.HandlePost)(w, r)
			return
		}
		w.WriteHeader(http.StatusMethodNotAllowed)
	})

	mux.HandleFunc("/api/signup", handlerPkg.HandleGenericPost(h.signupUC.Handle))
	mux.HandleFunc("/api/signup/app", handlerPkg.QueryHandler(h.viewSignupApplicationUC.Handle))

	mux.HandleFunc("/api/tickets/list-my-assignment", authMiddleware(handlerPkg.QueryHandler(h.listMyAssignmentUC.Handle)))

	mux.HandleFunc("/", handlerPkg.HandleStatic(h.webUC.RenderIndex))
}
