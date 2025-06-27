package submit

import (
	"context"
	"fmt"

	"github.com/jekiapp/topic-master/internal/model/acl"
	"github.com/tidwall/buntdb"
)

// Input struct for submitting an application
type SubmitApplicationRequest struct {
	EntityID        string   `json:"entity_id"`
	ApplicationType string   `json:"application_type"`
	Reason          string   `json:"reason"`
	Permission      []string `json:"permission"`
}

// Response struct for submitting an application
type SubmitApplicationResponse struct {
	AppID  string `json:"app_id"`
	AppURL string `json:"app_url"`
}

type SubmitApplicationUsecase struct {
	TopicActionUsecase TopicActionSubmitUsecase
}

func NewSubmitApplicationUsecase(db *buntdb.DB) SubmitApplicationUsecase {
	return SubmitApplicationUsecase{
		TopicActionUsecase: NewTopicActionSubmitUsecase(db),
	}
}

func (uc SubmitApplicationUsecase) Handle(ctx context.Context, req SubmitApplicationRequest) (SubmitApplicationResponse, error) {
	if req.ApplicationType == "" || len(req.Permission) == 0 {
		return SubmitApplicationResponse{}, fmt.Errorf("missing required fields")
	}

	if req.ApplicationType == acl.ApplicationType_TopicForm {
		return uc.TopicActionUsecase.Handle(ctx, req)
	}
	return SubmitApplicationResponse{}, fmt.Errorf("application type not supported")
}
