package main

import (
	"fmt"
	"log"
	"os"

	"github.com/tidwall/buntdb"
)

// This is a simple tool to help with db admin tasks
// create a func to truncate a table
// learn from internal/model and internal/repository
// give me example to truncate application table

// TruncateTable deletes all records in the application table.
func TruncateTable(db *buntdb.DB, prefix string) error {
	return db.Update(func(tx *buntdb.Tx) error {
		var keysToDelete []string
		tx.AscendKeys(prefix+"*", func(key, value string) bool {
			keysToDelete = append(keysToDelete, key)
			return true
		})
		for _, key := range keysToDelete {
			fmt.Println("deleting key: ", key)
			if _, err := tx.Delete(key); err != nil {
				return fmt.Errorf("failed to delete key %s: %w", key, err)
			}
		}
		return nil
	})
}

// example: go run main.go ../../data/topic-master.db truncate entity
func main() {
	if len(os.Args) < 2 {
		log.Fatalf("usage: %s <db_path>", os.Args[0])
	}
	dbPath := os.Args[1]
	db, err := buntdb.Open(dbPath)
	if err != nil {
		log.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	action := os.Args[2]

	if action == "truncate" {
		prefix := os.Args[3]
		if err := TruncateTable(db, prefix); err != nil {
			log.Fatalf("failed to truncate table: %v", err)
		}
		fmt.Println("Table truncated successfully.")
	}
}
