package logic

import (
	"github.com/jekiapp/topic-master/internal/config"
)

func Init(cfg *config.Config) error {
	// you can create various initialization in logic layer as needed
	// for example: callwrapper, featureflag, default value, etc.
	//
	// the pattern should follow as in repository init.go
	return nil
}
