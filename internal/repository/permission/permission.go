package permission

import (
	"fmt"
	"strings"

	"github.com/jekiapp/nsqper/internal/model/acl"
	"github.com/tidwall/buntdb"
	"github.com/vmihailenco/msgpack/v5"
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

func CreatePermission(db *buntdb.DB, permission acl.Permission) error {
	value, err := msgpack.Marshal(permission)
	if err != nil {
		return err
	}
	return db.Update(func(tx *buntdb.Tx) error {
		key := permission.GetPrimaryKey()
		_, replaced, err := tx.Set(key, string(value), nil)
		if err != nil {
			return err
		}
		if replaced {
			return fmt.Errorf("permission already exists")
		}
		for name, value := range permission.GetIndexValues() {
			_, _, err := tx.Set(key+":"+name, value, nil)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

func GetPermissionByNameEntity(db *buntdb.DB, name string, entityID string) (*acl.Permission, error) {
	permission := acl.Permission{
		Name:     name,
		EntityID: entityID,
	}

	pv := permission.GetIndexValues()["name_entity"]

	err := db.View(func(tx *buntdb.Tx) error {
		var foundKey string
		err := tx.AscendEqual(acl.IdxPermission_NameEntity, pv, func(key string, value string) bool {
			if pv == value {
				foundKey = strings.TrimSuffix(key, ":name_entity")
				return false
			}
			return true
		})

		if err != nil {
			return err
		}

		val, err := tx.Get(foundKey)
		if err != nil {
			return err
		}
		return msgpack.Unmarshal([]byte(val), &permission)
	})
	if err != nil {
		return nil, err
	}
	return &permission, nil
}

func GetPermissionByID(db *buntdb.DB, id string) (*acl.Permission, error) {
	var permission = acl.Permission{
		ID: id,
	}
	err := db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(permission.GetPrimaryKey())
		if err != nil {
			return err
		}
		return msgpack.Unmarshal([]byte(val), &permission)
	})
	if err != nil {
		return nil, err
	}
	return &permission, nil
}
