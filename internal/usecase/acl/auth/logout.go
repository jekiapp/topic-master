// implement logout usecase
// learn from login.go
// the logic will be:
// 1. delete the jwt token from the database
// 2. redirect to the login page

package acl

import (
	"net/http"
)

// LogoutUsecase handles user logout logic
// It clears the JWT cookie and redirects to the login page
// No DB access is needed as JWT is stateless

type LogoutUsecase struct{}

func NewLogoutUsecase() LogoutUsecase {
	return LogoutUsecase{}
}

func (uc LogoutUsecase) Handle(w http.ResponseWriter, r *http.Request) {
	// Clear the JWT cookie by setting it expired
	cookie := &http.Cookie{
		Name:     ACCESS_TOKEN_COOKIE_NAME,
		Value:    "",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
		MaxAge:   -1, // Expire immediately
	}
	http.SetCookie(w, cookie)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"redirect": "/#all-topics"}`))
}
