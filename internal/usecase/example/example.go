package example

import (
	"context"
	"database/sql"
	"net/http"

	"github.com/jekiapp/hi-mod-arch/internal/model"
	user_repo "github.com/jekiapp/hi-mod-arch/internal/repository/user"
)

type exampleUsecase struct {
	repo iExampleRepo
}

func NewExampleUsecase(dbCli *sql.DB,
	promotionCli, productCli, userCli *http.Client) exampleUsecase {
	return exampleUsecase{
		repo: exampleRepo{
			dbCli: dbCli,
		},
	}
}

type CheckoutPageRequest struct {
	UserID      int64
	PromoCoupon string
}

type CheckoutPageResponse struct {
	User       model.UserData
	Items      []model.CheckoutItem
	FinalPrice float64
}

func (uc exampleUsecase) HttpGenericHandler(ctx context.Context, input CheckoutPageRequest) (response CheckoutPageResponse, err error) {
	user, err := uc.repo.GetUserInfo(input.UserID)
	if err != nil {
		return response, err
	}

	return response, nil
}

//go:generate mockgen -source=render_page.go -destination=mock/render_page.go
type iExampleRepo interface {
	GetUserInfo(userID int64) (model.UserData, error)
}

type exampleRepo struct {
	dbCli *sql.DB
}

func (uc exampleRepo) GetUserInfo(userID int64) (model.UserData, error) {
	return user_repo.GetUserInfo(uc.dbCli, userID)
}
