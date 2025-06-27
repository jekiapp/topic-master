package mock

//go:generate go run github.com/golang/mock/mockgen -destination=signup_mock.go -package=mock github.com/jekiapp/topic-master/internal/usecase/acl/auth ISignupRepo
