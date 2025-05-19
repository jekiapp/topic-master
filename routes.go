package main

import (
	"net/http"

	"github.com/jekiapp/nsqper/internal/config"
	acl "github.com/jekiapp/nsqper/internal/usecase/acl"
	topicUC "github.com/jekiapp/nsqper/internal/usecase/topic"
	"github.com/jekiapp/nsqper/pkg/handler"
	handlerPkg "github.com/jekiapp/nsqper/pkg/handler"
	"github.com/tidwall/buntdb"
)

type Handler struct {
	createUserUC        acl.CreateUserUsecase
	loginUC             acl.LoginUsecase
	assignUserToGroupUC acl.AssignUserToGroupUsecase
	deleteUserUC        acl.DeleteUserUsecase
	createUserGroupUC   acl.CreateUserGroupUsecase
	createPermissionUC  acl.CreatePermissionUsecase
	changePasswordUC    acl.ChangePasswordUsecase
	syncTopicsUC        topicUC.SyncTopicsUsecase
}

func initHandler(db *buntdb.DB, cfg *config.Config) Handler {
	return Handler{
		createUserUC:        acl.NewCreateUserUsecase(db),
		loginUC:             acl.NewLoginUsecase(db, cfg),
		assignUserToGroupUC: acl.NewAssignUserToGroupUsecase(db),
		deleteUserUC:        acl.NewDeleteUserUsecase(db),
		createUserGroupUC:   acl.NewCreateUserGroupUsecase(db),
		createPermissionUC:  acl.NewCreatePermissionUsecase(db),
		changePasswordUC:    acl.NewChangePasswordUsecase(db),
		syncTopicsUC:        topicUC.NewSyncTopicsUsecase(db),
	}
}

func (h Handler) routes(mux *http.ServeMux) {
	mux.HandleFunc("/api/create-user", handlerPkg.HandleGenericPost(h.createUserUC.Handle))
	mux.HandleFunc("/api/login", handlerPkg.HandleGenericPost(h.loginUC.Handle))
	mux.HandleFunc("/api/assign-user-to-group", handlerPkg.HandleGenericPost(h.assignUserToGroupUC.Handle))
	mux.HandleFunc("/api/delete-user", handlerPkg.HandleGenericPost(h.deleteUserUC.Handle))
	mux.HandleFunc("/api/create-usergroup", handlerPkg.HandleGenericPost(h.createUserGroupUC.Handle))
	mux.HandleFunc("/api/create-permission", handlerPkg.HandleGenericPost(h.createPermissionUC.Handle))
	mux.HandleFunc("/api/change-password", handlerPkg.HandleGenericPost(h.changePasswordUC.Handle))
	mux.HandleFunc("/api/sync-topics", handler.QueryHandler(h.syncTopicsUC.HandleQuery))
}
