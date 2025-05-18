package db

import (
	"database/sql"
)

type DbConfig struct {
	Host string
	// ... etc
}

// TODO: Create our own DB object that implements new interface with sql.DB functions
// to abstract the infrastructure layer.
// then return that object, instead of *sql.DB object.
func InitDatabase(cfg DbConfig) (*sql.DB, error) {
	return &sql.DB{}, nil
}
