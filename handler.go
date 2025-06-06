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
	"github.com/jekiapp/topic-master/internal/usecase/tickets/action"
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
	listAllTopicsUC         topicUC.ListAllTopicsUsecase
	updateGroupByIDUC       aclGroup.UpdateGroupByIDUsecase
	deleteGroupUC           aclGroup.DeleteGroupUsecase
	resetPasswordUC         aclAuth.ResetPasswordUsecase
	signupUC                aclAuth.SignupUsecase
	viewSignupApplicationUC aclAuth.ViewSignupApplicationUsecase
	listMyAssignmentUC      tickets.ListMyAssignmentUsecase
	ticketDetailUC          tickets.TicketDetailUsecase
	actionCoordinatorUC     *action.ActionCoordinator
	getUsernameUC           aclAuth.GetUsernameUsecase
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
		listAllTopicsUC:         topicUC.NewListAllTopicsUsecase(db),
		updateGroupByIDUC:       aclGroup.NewUpdateGroupByIDUsecase(db),
		deleteGroupUC:           aclGroup.NewDeleteGroupUsecase(db),
		resetPasswordUC:         aclAuth.NewResetPasswordUsecase(db),
		signupUC:                aclAuth.NewSignupUsecase(db),
		viewSignupApplicationUC: aclAuth.NewViewSignupApplicationUsecase(db),
		listMyAssignmentUC:      tickets.NewListMyAssignmentUsecase(db),
		ticketDetailUC:          tickets.NewTicketDetailUsecase(db),
		actionCoordinatorUC:     action.NewActionCoordinator(db),
		getUsernameUC:           aclAuth.NewGetUsernameUsecase(),
	}
}

func (h Handler) routes(mux *http.ServeMux) {
	authMiddleware := handlerPkg.InitJWTMiddleware(string(h.config.SecretKey))
	rootMiddleware := handlerPkg.InitJWTMiddlewareWithRoot(string(h.config.SecretKey))

	mux.HandleFunc("/api/login", h.loginUC.Handle)
	mux.HandleFunc("/api/logout", h.logoutUC.Handle)

	mux.HandleFunc("/api/user/list", rootMiddleware(handlerPkg.HandleGenericPost(h.getUserListUC.Handle)))
	mux.HandleFunc("/api/user/create", rootMiddleware(handlerPkg.HandleGenericPost(h.createUserUC.Handle)))
	mux.HandleFunc("/api/user/update", rootMiddleware(handlerPkg.HandleGenericPost(h.updateUserUC.Handle)))
	mux.HandleFunc("/api/user/assign-to-group", rootMiddleware(handlerPkg.HandleGenericPost(h.assignUserToGroupUC.Handle)))
	mux.HandleFunc("/api/user/delete", rootMiddleware(handlerPkg.HandleGenericPost(h.deleteUserUC.Handle)))

	mux.HandleFunc("/api/create-permission", handlerPkg.HandleGenericPost(h.createPermissionUC.Handle))
	mux.HandleFunc("/api/change-password", handlerPkg.HandleGenericPost(h.changePasswordUC.Handle))
	mux.HandleFunc("/api/sync-topics", handlerPkg.HandleGenericGet(h.syncTopicsUC.HandleQuery))

	mux.HandleFunc("/api/group/create", rootMiddleware(handlerPkg.HandleGenericPost(h.createGroupUC.Handle)))
	mux.HandleFunc("/api/group/list", rootMiddleware(handlerPkg.HandleGenericPost(h.getGroupListUC.Handle)))
	mux.HandleFunc("/api/group/list-simple", handlerPkg.HandleGenericPost(h.getGroupListSimpleUC.Handle))
	mux.HandleFunc("/api/group/update-group-by-id", rootMiddleware(handlerPkg.HandleGenericPost(h.updateGroupByIDUC.Handle)))
	mux.HandleFunc("/api/group/delete-group", rootMiddleware(handlerPkg.HandleGenericPost(h.deleteGroupUC.Handle)))

	mux.HandleFunc("/api/topic/list-all-topics", handlerPkg.HandleGenericGet(h.listAllTopicsUC.HandleQuery))

	mux.HandleFunc("/api/reset-password", handlerPkg.HandleGetPost(
		h.resetPasswordUC.HandleGet,
		h.resetPasswordUC.HandlePost,
	))

	mux.HandleFunc("/api/signup", handlerPkg.HandleGenericPost(h.signupUC.Handle))
	mux.HandleFunc("/api/signup/app", handlerPkg.HandleGenericGet(h.viewSignupApplicationUC.Handle))

	mux.HandleFunc("/api/tickets/list-my-assignment", authMiddleware(handlerPkg.HandleGenericGet(h.listMyAssignmentUC.Handle)))
	mux.HandleFunc("/api/tickets/detail", authMiddleware(handlerPkg.HandleGenericGet(h.ticketDetailUC.Handle)))
	mux.HandleFunc("/api/tickets/action", authMiddleware(handlerPkg.HandleGenericPost(h.actionCoordinatorUC.Handle)))

	mux.HandleFunc("/api/user/get-username", authMiddleware(handlerPkg.HandleGenericGet(h.getUsernameUC.Handle)))

	mux.HandleFunc("/", handlerPkg.HandleStatic(h.webUC.RenderIndex))
}
