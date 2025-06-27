package mock

//go:generate mockgen -destination signup_mock.go -package mock github.com/jekiapp/topic-master/internal/usecase/acl/auth ISignupRepo
