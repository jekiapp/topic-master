package repository

import (
	"github.com/jekiapp/nsqper/internal/config"
	"github.com/jekiapp/nsqper/internal/model/acl"
	"github.com/jekiapp/nsqper/internal/repository/entity"
	"github.com/jekiapp/nsqper/internal/repository/lookupd"
	"github.com/jekiapp/nsqper/internal/repository/user"
	"github.com/tidwall/buntdb"
)

func Init(cfg *config.Config, db *buntdb.DB) error {
	lookupd.Init(cfg)
	err := entity.InitIndexEntity(db)
	if err != nil {
		return err
	}
	err = user.InitIndexUser(db)
	if err != nil {
		return err
	}
	err = user.InitIndexGroup(db)
	if err != nil {
		return err
	}
	err = user.InitIndexUserGroup(db)
	if err != nil {
		return err
	}
	return nil
}

func InitIndexResetPassword(db *buntdb.DB) error {
	indexes := acl.ResetPassword{}.GetIndexes()
	for _, index := range indexes {
		err := db.CreateIndex(index.Name, index.Pattern, index.Type)
		if err != nil {
			return err
		}
	}
	return nil
}
