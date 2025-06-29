package user

import (
	"context"
	"testing"

	acl "github.com/jekiapp/topic-master/internal/model/acl"
	"github.com/jekiapp/topic-master/pkg/util"
	"github.com/stretchr/testify/assert"
)

func TestGetUsernameUsecase_Handle(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		params  map[string]string
		wantErr bool
		want    GetUsernameResponse
	}{
		{
			name:    "user not authenticated",
			ctx:     context.Background(),
			params:  nil,
			wantErr: true,
		},
		{
			name: "success root",
			ctx: func() context.Context {
				user := &acl.User{ID: "u1", Username: "alice", Name: "Alice", Groups: []acl.GroupRole{{GroupName: acl.GroupRoot}}}
				return util.MockContextWithUser(context.Background(), user)
			}(),
			params: nil,
			want:   GetUsernameResponse{Name: "Alice", Username: "alice", Root: true, Groups: []string{acl.GroupRoot}},
		},
		{
			name: "success non-root single group",
			ctx: func() context.Context {
				user := &acl.User{ID: "u2", Username: "bob", Name: "Bob", Groups: []acl.GroupRole{{GroupName: "dev"}}}
				return util.MockContextWithUser(context.Background(), user)
			}(),
			params: nil,
			want:   GetUsernameResponse{Name: "Bob", Username: "bob", Root: false, Groups: []string{"dev"}},
		},
		{
			name: "success non-root multiple groups",
			ctx: func() context.Context {
				user := &acl.User{ID: "u3", Username: "carol", Name: "Carol", Groups: []acl.GroupRole{{GroupName: "dev"}, {GroupName: "ops"}}}
				return util.MockContextWithUser(context.Background(), user)
			}(),
			params: nil,
			want:   GetUsernameResponse{Name: "Carol", Username: "carol", Root: false, Groups: []string{"dev", "ops"}},
		},
		{
			name: "user with empty username",
			ctx: func() context.Context {
				user := &acl.User{ID: "u4", Username: "", Name: "Eve", Groups: []acl.GroupRole{{GroupName: "dev"}}}
				return util.MockContextWithUser(context.Background(), user)
			}(),
			params:  nil,
			wantErr: true,
		},
	}
	uc := GetUsernameUsecase{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := uc.Handle(tt.ctx, tt.params)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, resp)
			}
		})
	}
}
