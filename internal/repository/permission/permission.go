package permission

import (
	"github.com/jekiapp/topic-master/internal/model/acl"
	"github.com/jekiapp/topic-master/pkg/db"
	"github.com/tidwall/buntdb"
)

func InitIndexPermission(db *buntdb.DB) error {
	indexes := acl.Permission{}.GetIndexes()
	for _, index := range indexes {
		err := db.CreateIndex(index.Name, index.Pattern, index.Type)
		if err != nil {
			return err
		}
	}
	return nil
}

func CreatePermission(dbConn *buntdb.DB, permission acl.Permission) error {
	return db.Insert(dbConn, &permission)
}

func GetPermissionByNameEntity(dbConn *buntdb.DB, name string, entityID string) (*acl.Permission, error) {
	perm, err := db.SelectOne[acl.Permission](dbConn, name, acl.IdxPermission_NameEntity)
	if err != nil {
		return nil, err
	}
	return &perm, nil
}

func GetPermissionByID(dbConn *buntdb.DB, id string) (*acl.Permission, error) {
	perm, err := db.GetByID[acl.Permission](dbConn, id)
	if err != nil {
		return nil, err
	}
	return &perm, nil
}
