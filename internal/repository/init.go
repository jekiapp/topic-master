package repository

import (
	"github.com/jekiapp/nsqper/internal/config"
	"github.com/jekiapp/nsqper/internal/repository/lookupd"
)

func Init(cfg *config.Config) error {
	lookupd.Init(cfg)
	return nil
}
