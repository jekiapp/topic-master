package main

import (
	"net/http"

	"github.com/jekiapp/nsqper/internal/config"
	acl "github.com/jekiapp/nsqper/internal/usecase/acl"
	aclGroup "github.com/jekiapp/nsqper/internal/usecase/acl/group"
	aclUser "github.com/jekiapp/nsqper/internal/usecase/acl/user"
	topicUC "github.com/jekiapp/nsqper/internal/usecase/topic"
	webUC "github.com/jekiapp/nsqper/internal/usecase/web"
	handlerPkg "github.com/jekiapp/nsqper/pkg/handler"
	"github.com/tidwall/buntdb"
)

type Handler struct {
	config              *config.Config
	createUserUC        aclUser.CreateUserUsecase
	loginUC             acl.LoginUsecase
	assignUserToGroupUC acl.AssignUserToGroupUsecase
	deleteUserUC        aclUser.DeleteUserUsecase
	createGroupUC       aclGroup.CreateGroupUsecase
	createPermissionUC  acl.CreatePermissionUsecase
	changePasswordUC    aclUser.ChangePasswordUsecase
	syncTopicsUC        topicUC.SyncTopicsUsecase
	webUC               *webUC.WebUsecase
	getGroupListUC      aclGroup.GetGroupListUsecase
	getUserListUC       aclUser.GetUserListUsecase
	listTopicsUC        topicUC.ListTopicsUsecase
}

func initHandler(db *buntdb.DB, cfg *config.Config) Handler {
	webUsecase := webUC.NewWebUsecase()

	return Handler{
		config:              cfg,
		createUserUC:        aclUser.NewCreateUserUsecase(db),
		loginUC:             acl.NewLoginUsecase(db, cfg),
		assignUserToGroupUC: acl.NewAssignUserToGroupUsecase(db),
		deleteUserUC:        aclUser.NewDeleteUserUsecase(db),
		createGroupUC:       aclGroup.NewCreateGroupUsecase(db),
		createPermissionUC:  acl.NewCreatePermissionUsecase(db),
		changePasswordUC:    aclUser.NewChangePasswordUsecase(db),
		syncTopicsUC:        topicUC.NewSyncTopicsUsecase(db),
		webUC:               webUsecase,
		getGroupListUC:      aclGroup.NewGetGroupListUsecase(db),
		getUserListUC:       aclUser.NewGetUserListUsecase(db),
		listTopicsUC:        topicUC.NewListTopicsUsecase(db),
	}
}

func (h Handler) routes(mux *http.ServeMux) {
	authMiddleware := handlerPkg.InitJWTMiddleware(string(h.config.SecretKey))

	mux.HandleFunc("/api/login", h.loginUC.Handle)
	mux.HandleFunc("/api/create-user", handlerPkg.HandleGenericPost(h.createUserUC.Handle))
	mux.HandleFunc("/api/assign-user-to-group", handlerPkg.HandleGenericPost(h.assignUserToGroupUC.Handle))
	mux.HandleFunc("/api/delete-user", handlerPkg.HandleGenericPost(h.deleteUserUC.Handle))
	mux.HandleFunc("/api/create-group", handlerPkg.HandleGenericPost(h.createGroupUC.Handle))
	mux.HandleFunc("/api/create-permission", handlerPkg.HandleGenericPost(h.createPermissionUC.Handle))
	mux.HandleFunc("/api/change-password", handlerPkg.HandleGenericPost(h.changePasswordUC.Handle))
	mux.HandleFunc("/api/sync-topics", handlerPkg.QueryHandler(h.syncTopicsUC.HandleQuery))

	mux.HandleFunc("/api/group-list", handlerPkg.HandleGenericPost(h.getGroupListUC.Handle))
	mux.HandleFunc("/api/user-list", handlerPkg.HandleGenericPost(h.getUserListUC.Handle))

	mux.HandleFunc("/api/topic/list-topics", authMiddleware(handlerPkg.QueryHandler(h.listTopicsUC.HandleQuery)))

	mux.HandleFunc("/", handlerPkg.HandleStatic(h.webUC.RenderIndex))
}
