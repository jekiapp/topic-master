package db

import (
	"fmt"
	"log"
	"strings"

	"github.com/tidwall/buntdb"
	"github.com/vmihailenco/msgpack/v5"
)

var ErrNotFound = buntdb.ErrNotFound

type DbConfig struct {
	Host string
}

type Index struct {
	Name    string
	Pattern string
	Type    IndexType
}

type IndexType func(a, b string) bool

type IndexItem struct {
	Name string
	Type IndexType
}

type Table interface {
	GetIndexes() []Index
}

type GetRecordByIndexes interface {
	GetPrimaryKey
	GetIndexValues() map[string]string
}

type GetPrimaryKey interface {
	GetPrimaryKey(id string) string
}

type DeleteRecordByID interface {
	GetPrimaryKey
	Table
}

type Pagination struct {
	Page  int // 1-based
	Limit int
}

func Insert(db *buntdb.DB, record GetRecordByIndexes) error {
	return db.Update(func(tx *buntdb.Tx) error {
		key := record.GetPrimaryKey("")
		id := strings.Split(key, ":")[1]
		if id == "" {
			return fmt.Errorf("id is empty")
		}
		msgpackValue, err := msgpack.Marshal(record)
		if err != nil {
			return err
		}

		_, err = tx.Get(key)
		if err == nil {
			return fmt.Errorf("data %s already exists", key)
		}

		// set primary data
		value := string(msgpackValue)
		_, _, err = tx.Set(key, value, nil)
		if err != nil {
			return err
		}

		// set index with rollback on failure
		setIndexes := make([]string, 0)
		for name, value := range record.GetIndexValues() {
			if value == "" {
				return fmt.Errorf("%s: index %s value is empty", key, name)
			}
			idxKey := key + ":" + name
			_, _, err = tx.Set(idxKey, value, nil)
			if err != nil {
				log.Printf("error setting index: %s", err)
				// rollback: delete primary and any set indexes
				tx.Delete(key)
				for _, sIdxKey := range setIndexes {
					tx.Delete(sIdxKey)
				}
				return fmt.Errorf("failed to set index %s: %w", name, err)
			}
			setIndexes = append(setIndexes, idxKey)
		}
		return nil
	})
}

func Update(db *buntdb.DB, record GetRecordByIndexes) error {
	return db.Update(func(tx *buntdb.Tx) error {
		key := record.GetPrimaryKey("")
		msgpackValue, err := msgpack.Marshal(record)
		if err != nil {
			return err
		}

		_, err = tx.Get(key)
		if err != nil {
			return fmt.Errorf("data %s not found", key)
		}

		value := string(msgpackValue)
		prevValue, _, err := tx.Set(key, value, nil)
		if err != nil {
			return err
		}

		// set index with rollback on failure
		setIndexes := make([]string, 0)
		for name, value := range record.GetIndexValues() {
			idxKey := key + ":" + name
			_, _, err = tx.Set(idxKey, value, nil)
			if err != nil {
				log.Printf("error setting index: %s", err)
				// rollback: restore previous value and remove set indexes
				tx.Set(key, prevValue, nil)
				for _, sIdxKey := range setIndexes {
					tx.Delete(sIdxKey)
				}
				return fmt.Errorf("failed to set index %s: %w", name, err)
			}
			setIndexes = append(setIndexes, idxKey)
		}
		return nil
	})
}

func Upsert(db *buntdb.DB, record GetRecordByIndexes) error {
	return db.Update(func(tx *buntdb.Tx) error {
		key := record.GetPrimaryKey("")
		msgpackValue, err := msgpack.Marshal(record)
		if err != nil {
			return err
		}

		value := string(msgpackValue)
		prevValue, replaced, err := tx.Set(key, value, nil)
		if err != nil {
			return err
		}

		// set index with rollback on failure
		setIndexes := make([]string, 0)
		for name, value := range record.GetIndexValues() {
			idxKey := key + ":" + name
			_, _, err = tx.Set(idxKey, value, nil)
			if err != nil {
				log.Printf("error setting index: %s", err)
				// rollback: restore or delete primary, remove set indexes
				if replaced {
					tx.Set(key, prevValue, nil)
				} else {
					tx.Delete(key)
				}
				for _, sIdxKey := range setIndexes {
					tx.Delete(sIdxKey)
				}
				return fmt.Errorf("failed to set index %s: %w", name, err)
			}
			setIndexes = append(setIndexes, idxKey)
		}
		return nil
	})
}

type processFunc func(key string, value string) bool

func SelectAll[T any](db *buntdb.DB, pivot string, indexName string) ([]T, error) {
	return SelectPaginated[T](db, pivot, indexName, nil)
}

func SelectPaginated[T any](db *buntdb.DB, pivot string, indexName string, pagination *Pagination) ([]T, error) {
	var results []T

	// get index field name
	// tablename:indexname
	idxField := strings.Split(indexName, ":")[1]

	var skip, limit int
	if pagination != nil && pagination.Limit > 0 {
		limit = pagination.Limit
		if pagination.Page == 0 {
			pagination.Page = 1
		}
		if pagination.Page > 1 {
			skip = (pagination.Page - 1) * pagination.Limit
		}
	}
	idx := 0
	collected := 0

	// handle composite index value
	parts := strings.Split(pivot, ":")
	prefix := ""
	if len(parts) > 1 {
		// get all the strings until the last ":"
		parts = parts[:len(parts)-1]
		// join the parts with ":"
		prefix = strings.Join(parts, ":")
	}
	rangeOp := false

	process := func(tx *buntdb.Tx) processFunc {
		return func(key string, value string) bool {
			if rangeOp && prefix != "" {
				if !strings.HasPrefix(value, prefix) {
					return false
				}
			}
			if skip > 0 && idx < skip {
				idx++
				return true
			}
			if limit > 0 && collected >= limit {
				return false // stop iteration
			}
			newKey := strings.TrimSuffix(key, ":"+idxField)
			val, err := tx.Get(newKey)
			if err != nil {
				log.Printf("error getting value: %s", err)
				return false
			}
			var result T
			err = msgpack.Unmarshal([]byte(val), &result)
			if err != nil {
				log.Printf("error unmarshalling value: %s", err)
				return false
			}
			results = append(results, result)
			idx++
			collected++
			return true
		}
	}

	err := db.View(func(tx *buntdb.Tx) error {
		if pivot == "*" {
			tx.Ascend(indexName, process(tx))
			return nil
		}

		if strings.HasPrefix(pivot, "=") {
			pivot = strings.TrimPrefix(pivot, "=")
			tx.AscendEqual(indexName, pivot, process(tx))
			return nil
		}

		if strings.HasPrefix(pivot, "-<=") {
			pivot = strings.TrimPrefix(pivot, "-<=")
			prefix = strings.TrimPrefix(prefix, "-<=")
			rangeOp = true
			tx.DescendLessOrEqual(indexName, pivot, process(tx))
			return nil
		}

		if strings.HasPrefix(pivot, ">=") {
			pivot = strings.TrimPrefix(pivot, ">=")
			prefix = strings.TrimPrefix(prefix, ">=")
			rangeOp = true
			tx.AscendGreaterOrEqual(indexName, pivot, process(tx))
			return nil
		}

		if strings.HasPrefix(pivot, "<") {
			pivot = strings.TrimPrefix(pivot, "<")
			prefix = strings.TrimPrefix(prefix, "<")
			rangeOp = true
			tx.AscendLessThan(indexName, pivot, process(tx))
			return nil
		}

		return fmt.Errorf("pivot %s is not valid", pivot)
	})
	if err != nil {
		return nil, err
	}
	if len(results) == 0 {
		return results, ErrNotFound
	}
	return results, nil
}

func SelectOne[Y any](db *buntdb.DB, pivot string, indexName string) (Y, error) {
	result := new(Y)

	// get index field name
	// tablename:indexname
	idxField := strings.Split(indexName, ":")[1]

	if pivot == "" {
		return *result, fmt.Errorf("pivot %s is empty", idxField)
	}

	found := false
	err := db.View(func(tx *buntdb.Tx) error {
		tx.AscendEqual(indexName, pivot, func(key string, value string) bool {
			// they key would be like this:
			// table:<id>:indexname
			// pick just table:<id>
			newKey := strings.TrimSuffix(key, ":"+idxField)
			val, err := tx.Get(newKey)
			if err != nil {
				log.Printf("error getting value: %s", err)
				return false
			}
			err = msgpack.Unmarshal([]byte(val), &result)
			if err != nil {
				log.Printf("error unmarshalling value: %s", err)
				return false
			}
			found = true
			return false
		})
		return nil
	})
	if !found {
		return *result, ErrNotFound
	}
	return *result, err
}

// make sure the id in the primary key is not empty
func GetByID[Y any](db *buntdb.DB, id string) (Y, error) {
	result := new(Y)
	rec, ok := any(result).(GetPrimaryKey)
	if !ok {
		return *result, fmt.Errorf("type %T is not implement GetPrimaryKey interface", result)
	}

	key := rec.GetPrimaryKey(id)
	if id == "" {
		return *result, fmt.Errorf("id is empty")
	}

	err := db.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(key)
		if err != nil {
			return err
		}
		return msgpack.Unmarshal([]byte(val), result)
	})
	return *result, err
}

func DeleteByID[Y any](db *buntdb.DB, id string) error {
	result := new(Y)
	rec, ok := any(result).(DeleteRecordByID)
	if !ok {
		return fmt.Errorf("type %T is not implement DeleteRecordByID interface", result)
	}

	return db.Update(func(tx *buntdb.Tx) error {
		key := rec.GetPrimaryKey(id)
		_, err := tx.Delete(key)
		if err != nil {
			return fmt.Errorf("error deleting primary key: %s, %w", key, err)
		}
		// delete indexes
		for _, index := range rec.GetIndexes() {
			field := strings.Split(index.Name, ":")[1]
			idxKey := key + ":" + field
			_, err := tx.Delete(idxKey)
			if err != nil {
				return fmt.Errorf("error deleting index key: %s, %w", idxKey, err)
			}
		}
		return nil
	})
}

func DeleteByIndex(db *buntdb.DB, record GetRecordByIndexes, indexName string) error {
	return db.Update(func(tx *buntdb.Tx) error {
		idxField := strings.Split(indexName, ":")[1]
		pivot := record.GetIndexValues()[idxField]
		if pivot == "" {
			return fmt.Errorf("pivot %s is empty", idxField)
		}

		var primaryKeys []string
		var indexKeys []string
		tx.AscendEqual(indexName, pivot, func(key string, value string) bool {
			pk := strings.TrimSuffix(key, ":"+idxField)
			primaryKeys = append(primaryKeys, pk)

			// delete indexes
			for name := range record.GetIndexValues() {
				idxKey := pk + ":" + name
				indexKeys = append(indexKeys, idxKey)
			}
			return true
		})

		for _, pk := range primaryKeys {
			_, err := tx.Delete(pk)
			if err != nil {
				return fmt.Errorf("error deleting primary key: %s, %w", pk, err)
			}
		}
		for _, idxKey := range indexKeys {
			_, err := tx.Delete(idxKey)
			if err != nil {
				return fmt.Errorf("error deleting index key: %s, %w", idxKey, err)
			}
		}
		return nil
	})
}
