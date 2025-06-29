package topic

import (
	"context"
	"errors"
	"testing"

	"github.com/jekiapp/topic-master/internal/model"
	"github.com/jekiapp/topic-master/internal/model/acl"
	entity "github.com/jekiapp/topic-master/internal/model/entity"
	topic_mock "github.com/jekiapp/topic-master/internal/usecase/topic/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func ctxWithUserIDMyTopics(userID string) context.Context {
	if userID == "" {
		return context.Background()
	}
	claims := &acl.JWTClaims{UserID: userID}
	return context.WithValue(context.Background(), model.UserInfoKey, claims)
}

func TestListMyBookmarkedTopicsUsecase_HandleQuery(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name      string
		mockSetup func(m *topic_mock.MockiListMyBookmarkedTopicsRepo)
		userID    string
		wantErr   bool
		wantLen   int
	}{
		{
			name:      "no user",
			mockSetup: func(m *topic_mock.MockiListMyBookmarkedTopicsRepo) {},
			userID:    "",
			wantErr:   false,
			wantLen:   0,
		},
		{
			name: "repo error",
			mockSetup: func(m *topic_mock.MockiListMyBookmarkedTopicsRepo) {
				m.EXPECT().ListBookmarkedTopicIDsByUser("alice").Return(nil, errors.New("db error"))
			},
			userID:  "alice",
			wantErr: true,
			wantLen: 0,
		},
		{
			name: "no bookmarks",
			mockSetup: func(m *topic_mock.MockiListMyBookmarkedTopicsRepo) {
				m.EXPECT().ListBookmarkedTopicIDsByUser("bob").Return([]string{}, nil)
			},
			userID:  "bob",
			wantErr: false,
			wantLen: 0,
		},
		{
			name: "topics found",
			mockSetup: func(m *topic_mock.MockiListMyBookmarkedTopicsRepo) {
				ids := []string{"topicA"}
				entities := []entity.Entity{{ID: "topicA", Name: "Topic Alpha", Description: "descA", GroupOwner: "groupAlpha"}}
				m.EXPECT().ListBookmarkedTopicIDsByUser("alice").Return(ids, nil)
				m.EXPECT().GetNsqTopicEntitiesByIDs(ids).Return(entities, nil)
			},
			userID:  "alice",
			wantErr: false,
			wantLen: 1,
		},
		{
			name: "ListBookmarkedTopicIDsByUser returns nil, no error",
			mockSetup: func(m *topic_mock.MockiListMyBookmarkedTopicsRepo) {
				m.EXPECT().ListBookmarkedTopicIDsByUser("bob").Return(nil, nil)
			},
			userID:  "bob",
			wantErr: false,
			wantLen: 0,
		},
		{
			name: "GetNsqTopicEntitiesByIDs returns error",
			mockSetup: func(m *topic_mock.MockiListMyBookmarkedTopicsRepo) {
				ids := []string{"topicA"}
				m.EXPECT().ListBookmarkedTopicIDsByUser("alice").Return(ids, nil)
				m.EXPECT().GetNsqTopicEntitiesByIDs(ids).Return(nil, errors.New("err"))
			},
			userID:  "alice",
			wantErr: true,
			wantLen: 0,
		},
		{
			name: "multiple topic IDs/entities",
			mockSetup: func(m *topic_mock.MockiListMyBookmarkedTopicsRepo) {
				ids := []string{"topicA", "topicB"}
				entities := []entity.Entity{{ID: "topicA", Name: "Topic Alpha"}, {ID: "topicB", Name: "Topic Beta"}}
				m.EXPECT().ListBookmarkedTopicIDsByUser("bob").Return(ids, nil)
				m.EXPECT().GetNsqTopicEntitiesByIDs(ids).Return(entities, nil)
			},
			userID:  "bob",
			wantErr: false,
			wantLen: 2,
		},
		{
			name:      "user present but ID empty",
			mockSetup: func(m *topic_mock.MockiListMyBookmarkedTopicsRepo) {},
			userID:    "",
			wantErr:   false,
			wantLen:   0,
		},
		{
			name: "ListBookmarkedTopicIDsByUser returns error",
			mockSetup: func(m *topic_mock.MockiListMyBookmarkedTopicsRepo) {
				m.EXPECT().ListBookmarkedTopicIDsByUser("alice").Return(nil, errors.New("err"))
			},
			userID:  "alice",
			wantErr: true,
			wantLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := topic_mock.NewMockiListMyBookmarkedTopicsRepo(ctrl)
			tt.mockSetup(mockRepo)
			uc := ListMyBookmarkedTopicsUsecase{repo: mockRepo}
			ctx := ctxWithUserIDMyTopics(tt.userID)
			resp, err := uc.HandleQuery(ctx, nil)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, resp.Topics, tt.wantLen)
			}
		})
	}
}
