package topic

import (
	"context"
	"errors"
	"testing"

	"github.com/jekiapp/topic-master/internal/model"
	"github.com/jekiapp/topic-master/internal/model/acl"
	entity "github.com/jekiapp/topic-master/internal/model/entity"
	topic_mock "github.com/jekiapp/topic-master/internal/usecase/topic/mock"
	dbPkg "github.com/jekiapp/topic-master/pkg/db"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func ctxWithUserID(userID string) context.Context {
	if userID == "" {
		return context.Background()
	}
	claims := &acl.JWTClaims{UserID: userID}
	return context.WithValue(context.Background(), model.UserInfoKey, claims)
}

func TestListAllTopicsUsecase_HandleQuery(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name      string
		mockSetup func(m *topic_mock.MockiListTopicsRepo)
		params    map[string]string
		userID    string
		wantErr   bool
		wantLen   int
	}{
		{
			name: "repo error",
			mockSetup: func(m *topic_mock.MockiListTopicsRepo) {
				m.EXPECT().GetAllNsqTopicEntities().Return(nil, errors.New("db error"))
			},
			params:  map[string]string{},
			userID:  "alice",
			wantErr: true,
			wantLen: 0,
		},
		{
			name: "empty topics",
			mockSetup: func(m *topic_mock.MockiListTopicsRepo) {
				m.EXPECT().GetAllNsqTopicEntities().Return([]entity.Entity{}, nil)
			},
			params:  map[string]string{},
			userID:  "alice",
			wantErr: false,
			wantLen: 0,
		},
		{
			name: "topics with bookmarks",
			mockSetup: func(m *topic_mock.MockiListTopicsRepo) {
				entities := []entity.Entity{{ID: "topicA", Name: "Topic Alpha", Description: "descA", GroupOwner: "groupAlpha"}}
				m.EXPECT().GetAllNsqTopicEntities().Return(entities, nil)
				m.EXPECT().IsBookmarked("topicA", "alice").Return(true, nil)
			},
			params:  map[string]string{},
			userID:  "alice",
			wantErr: false,
			wantLen: 1,
		},
		{
			name: "dbPkg.ErrNotFound returns empty list",
			mockSetup: func(m *topic_mock.MockiListTopicsRepo) {
				m.EXPECT().GetAllNsqTopicEntities().Return(nil, dbPkg.ErrNotFound)
			},
			params:  map[string]string{},
			userID:  "bob",
			wantErr: false,
			wantLen: 0,
		},
		{
			name: "multiple topics, mixed bookmarks",
			mockSetup: func(m *topic_mock.MockiListTopicsRepo) {
				entities := []entity.Entity{{ID: "topicA", Name: "Topic Alpha"}, {ID: "topicB", Name: "Topic Beta"}}
				m.EXPECT().GetAllNsqTopicEntities().Return(entities, nil)
				m.EXPECT().IsBookmarked("topicA", "bob").Return(true, nil)
				m.EXPECT().IsBookmarked("topicB", "bob").Return(false, nil)
			},
			params:  map[string]string{},
			userID:  "bob",
			wantErr: false,
			wantLen: 2,
		},
		{
			name: "IsBookmarked returns error, should default to false",
			mockSetup: func(m *topic_mock.MockiListTopicsRepo) {
				entities := []entity.Entity{{ID: "topicA", Name: "Topic Alpha"}}
				m.EXPECT().GetAllNsqTopicEntities().Return(entities, nil)
				m.EXPECT().IsBookmarked("topicA", "alice").Return(false, errors.New("err"))
			},
			params:  map[string]string{},
			userID:  "alice",
			wantErr: false,
			wantLen: 1,
		},
		{
			name: "user present but ID empty",
			mockSetup: func(m *topic_mock.MockiListTopicsRepo) {
				entities := []entity.Entity{{ID: "topicA", Name: "Topic Alpha"}}
				m.EXPECT().GetAllNsqTopicEntities().Return(entities, nil)
			},
			params:  map[string]string{},
			userID:  "",
			wantErr: false,
			wantLen: 1,
		},
		{
			name: "no topics returned",
			mockSetup: func(m *topic_mock.MockiListTopicsRepo) {
				m.EXPECT().GetAllNsqTopicEntities().Return([]entity.Entity{}, nil)
			},
			params:  map[string]string{},
			userID:  "bob",
			wantErr: false,
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := topic_mock.NewMockiListTopicsRepo(ctrl)
			tt.mockSetup(mockRepo)
			uc := ListAllTopicsUsecase{repo: mockRepo}
			ctx := ctxWithUserID(tt.userID)
			resp, err := uc.HandleQuery(ctx, tt.params)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, resp.Topics, tt.wantLen)
			}
		})
	}
}
