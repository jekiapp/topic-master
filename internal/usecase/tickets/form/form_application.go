package form

import (
	"context"
	"errors"

	"github.com/jekiapp/topic-master/internal/model/acl"
	"github.com/tidwall/buntdb"
)

type NewApplicationRequest struct {
	EntityID string `json:"entity_id"`
	Type     string `json:"type"`
}

type NewApplicationResponse struct {
	Title       string             `json:"title"`
	Applicant   applicantResponse  `json:"applicant"`
	Type        string             `json:"type"` // topic, group, user
	Reviewers   []reviewerResponse `json:"reviewers"`
	Fields      []fieldResponse    `json:"fields"`
	Permissions []acl.Permission   `json:"permissions"`
}

type applicantResponse struct {
	Username string `json:"username"`
	Name     string `json:"name"`
}

type reviewerResponse struct {
	Username string `json:"username"`
	Name     string `json:"name"`
}

type fieldResponse struct {
	Label        string `json:"label"`
	Type         string `json:"type"` // text, bool, textarea, hidden
	Required     bool   `json:"required"`
	DefaultValue string `json:"default_value"`
	Editable     bool   `json:"editable"`
}

type NewApplicationUsecase struct {
	topicActionUsecase TopicActionUsecase
}

func NewNewApplicationUsecase(db *buntdb.DB) NewApplicationUsecase {
	return NewApplicationUsecase{
		topicActionUsecase: NewFormTopicActionUsecase(db),
	}
}

func (uc NewApplicationUsecase) Handle(ctx context.Context, req map[string]string) (NewApplicationResponse, error) {
	entityID := req["entity_id"]
	typeApplication := req["type"]

	if typeApplication == acl.ApplicationType_TopicForm {
		return uc.topicActionUsecase.getTopicForm(ctx, entityID)
	}

	return NewApplicationResponse{}, errors.New("type not supported")
}
