package acl

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogoutUsecase_Handle(t *testing.T) {
	uc := NewLogoutUsecase()

	tests := []struct {
		name string
	}{
		{
			name: "logout clears cookie and redirects",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/logout", nil)
			rw := httptest.NewRecorder()
			uc.Handle(rw, req)

			// Check cookie
			cookies := rw.Result().Cookies()
			var found bool
			for _, c := range cookies {
				if c.Name == ACCESS_TOKEN_COOKIE_NAME {
					assert.Equal(t, "", c.Value)
					assert.True(t, c.MaxAge < 0)
					found = true
				}
			}
			assert.True(t, found, "access token cookie should be set")

			// Check redirect
			assert.Equal(t, http.StatusSeeOther, rw.Result().StatusCode)
			assert.Equal(t, "/", rw.Result().Header.Get("Location"))
		})
	}
}
