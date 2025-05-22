package main

import (
	"net/http"

	"github.com/jekiapp/nsqper/internal/config"
	acl "github.com/jekiapp/nsqper/internal/usecase/acl"
	topicUC "github.com/jekiapp/nsqper/internal/usecase/topic"
	webUC "github.com/jekiapp/nsqper/internal/usecase/web"
	"github.com/jekiapp/nsqper/pkg/handler"
	handlerPkg "github.com/jekiapp/nsqper/pkg/handler"
	"github.com/tidwall/buntdb"
)

type Handler struct {
	config              *config.Config
	createUserUC        acl.CreateUserUsecase
	loginUC             acl.LoginUsecase
	assignUserToGroupUC acl.AssignUserToGroupUsecase
	deleteUserUC        acl.DeleteUserUsecase
	createUserGroupUC   acl.CreateUserGroupUsecase
	createPermissionUC  acl.CreatePermissionUsecase
	changePasswordUC    acl.ChangePasswordUsecase
	syncTopicsUC        topicUC.SyncTopicsUsecase
	webUC               *webUC.WebUsecase
}

func initHandler(db *buntdb.DB, cfg *config.Config) Handler {
	webUsecase := webUC.NewWebUsecase()

	return Handler{
		config:              cfg,
		createUserUC:        acl.NewCreateUserUsecase(db),
		loginUC:             acl.NewLoginUsecase(db, cfg),
		assignUserToGroupUC: acl.NewAssignUserToGroupUsecase(db),
		deleteUserUC:        acl.NewDeleteUserUsecase(db),
		createUserGroupUC:   acl.NewCreateUserGroupUsecase(db),
		createPermissionUC:  acl.NewCreatePermissionUsecase(db),
		changePasswordUC:    acl.NewChangePasswordUsecase(db),
		syncTopicsUC:        topicUC.NewSyncTopicsUsecase(db),
		webUC:               webUsecase,
	}
}

func (h Handler) routes(mux *http.ServeMux) {
	// authMiddleware := handlerPkg.InitJWTMiddleware(string(h.config.SecretKey))

	mux.HandleFunc("/api/create-user", handlerPkg.HandleGenericPost(h.createUserUC.Handle))
	mux.HandleFunc("/api/login", h.loginUC.Handle)
	mux.HandleFunc("/api/assign-user-to-group", handlerPkg.HandleGenericPost(h.assignUserToGroupUC.Handle))
	mux.HandleFunc("/api/delete-user", handlerPkg.HandleGenericPost(h.deleteUserUC.Handle))
	mux.HandleFunc("/api/create-usergroup", handlerPkg.HandleGenericPost(h.createUserGroupUC.Handle))
	mux.HandleFunc("/api/create-permission", handlerPkg.HandleGenericPost(h.createPermissionUC.Handle))
	mux.HandleFunc("/api/change-password", handlerPkg.HandleGenericPost(h.changePasswordUC.Handle))
	mux.HandleFunc("/api/sync-topics", handler.QueryHandler(h.syncTopicsUC.HandleQuery))

	mux.HandleFunc("/login", handlerPkg.HandleStatic(h.webUC.RenderIndex))
	mux.HandleFunc("/login/*", handlerPkg.HandleStatic(h.webUC.RenderIndex))
	mux.HandleFunc("/", handlerPkg.HandleStatic(h.webUC.RenderIndex))
}
