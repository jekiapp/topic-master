package permission

import (
	"github.com/jekiapp/nsqper/internal/model/acl"
	"github.com/jekiapp/nsqper/pkg/db"
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
	return db.Insert(dbConn, permission)
}

func GetPermissionByNameEntity(dbConn *buntdb.DB, name string, entityID string) (*acl.Permission, error) {
	tmp := acl.Permission{Name: name, EntityID: entityID}
	perm, err := db.SelectOne[acl.Permission](dbConn, tmp, acl.IdxPermission_NameEntity)
	if err != nil {
		return nil, err
	}
	return &perm, nil
}

func GetPermissionByID(dbConn *buntdb.DB, id string) (*acl.Permission, error) {
	tmp := acl.Permission{ID: id}
	perm, err := db.GetByID[acl.Permission](dbConn, tmp)
	if err != nil {
		return nil, err
	}
	return &perm, nil
}
