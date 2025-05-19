package topic

import (
	"errors"

	"github.com/jekiapp/nsqper/internal/model/acl"
	"github.com/tidwall/buntdb"
)

type iSyncTopics interface {
	GetAllTopics() ([]string, error)
	GetNsqTopicEntity(topic string) (*acl.Entity, error)
	CreateNsqTopicEntity(topic string) (*acl.Entity, error)
}

func SyncTopics(db *buntdb.DB, iSyncTopics iSyncTopics) error {
	topics, err := iSyncTopics.GetAllTopics()
	if err != nil {
		return err
	}
	var errs []error
	for _, topic := range topics {
		entity, err := iSyncTopics.GetNsqTopicEntity(topic)
		if err != nil && err.Error() != "not found" {
			errs = append(errs, errors.New("GetNsqTopicEntity("+topic+"): "+err.Error()))
			continue
		}
		if entity == nil {
			_, createErr := iSyncTopics.CreateNsqTopicEntity(topic)
			if createErr != nil {
				errs = append(errs, errors.New("CreateNsqTopicEntity("+topic+"): "+createErr.Error()))
			}
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}
