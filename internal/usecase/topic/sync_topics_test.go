package topic

import (
	"context"
	"errors"
	"testing"

	topic_mock "github.com/jekiapp/topic-master/internal/usecase/topic/mock"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestSyncTopicsUsecase_HandleQuery(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	tests := []struct {
		name      string
		mockSetup func(m *topic_mock.MockiSyncTopicsRepo)
		wantErr   bool
		wantSucc  bool
	}{
		{
			name: "sync error",
			mockSetup: func(m *topic_mock.MockiSyncTopicsRepo) {
				m.EXPECT().GetAllTopics().Return(nil, errors.New("sync error"))
			},
			wantErr:  true,
			wantSucc: false,
		},
		{
			name: "sync success",
			mockSetup: func(m *topic_mock.MockiSyncTopicsRepo) {
				m.EXPECT().GetAllTopics().Return([]string{"t1"}, nil)
			},
			wantErr:  false,
			wantSucc: true,
		},
		{
			name: "GetAllTopics returns empty slice",
			mockSetup: func(m *topic_mock.MockiSyncTopicsRepo) {
				m.EXPECT().GetAllTopics().Return([]string{}, nil)
			},
			wantErr:  false,
			wantSucc: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRepo := topic_mock.NewMockiSyncTopicsRepo(ctrl)
			tt.mockSetup(mockRepo)
			uc := SyncTopicsUsecase{repo: mockRepo}
			resp, err := uc.HandleQuery(context.Background(), nil)
			if tt.wantErr {
				assert.Error(t, err)
				assert.False(t, resp.Success)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantSucc, resp.Success)
			}
		})
	}
}
