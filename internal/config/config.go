package config

import (
	"crypto/rand"
	"encoding/base64"
	"errors"

	"github.com/tidwall/buntdb"
	"github.com/vmihailenco/msgpack/v5"

	// Correct internal imports

	usergroup "github.com/jekiapp/nsqper/internal/repository/user"
	dbPkg "github.com/jekiapp/nsqper/pkg/db"
)

const configKey = "nsqper_config"

type Config struct {
	NSQLookupdHTTPAddr string
	NSQDAddr           string
	SecretKey          []byte
}

func NewConfig(db *buntdb.DB) (*Config, error) {
	var cfg Config
	var raw string
	// Fetch the msgpack-encoded config from buntDB
	err := db.View(func(tx *buntdb.Tx) error {
		var err error
		raw, err = tx.Get(configKey)
		if err == buntdb.ErrNotFound {
			return errors.New("nsqper_config not found in buntDB")
		} else if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Unmarshal using msgpack
	if err := msgpack.Unmarshal([]byte(raw), &cfg); err != nil {
		return nil, errors.New("failed to unmarshal nsqper_config: " + err.Error())
	}

	return &cfg, nil
}

// CheckRootGroupAndUserExist returns true if both root group and root user exist in the DB.
func CheckRootGroupAndUserExist(db *buntdb.DB) (bool, error) {
	_, err := usergroup.GetGroupByName(db, "root")
	if err != nil && err != dbPkg.ErrNotFound {
		return false, err
	}
	if err == dbPkg.ErrNotFound {
		return false, nil
	}
	_, err = usergroup.GetUserByUsername(db, "root")
	if err != nil && err != dbPkg.ErrNotFound {
		return false, err
	}
	if err == dbPkg.ErrNotFound {
		return false, nil
	}
	return true, nil
}

// SetupNewConfig creates a new config with a random secret key, saves it to the db, and returns it.
func SetupNewConfig(db *buntdb.DB, nsqlookupdHTTPAddr string) (*Config, error) {
	// Generate a random 32-byte secret key
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return nil, errors.New("failed to generate random secret key: " + err.Error())
	}
	secretKey := []byte(base64.StdEncoding.EncodeToString(key))

	cfg := &Config{
		NSQLookupdHTTPAddr: nsqlookupdHTTPAddr,
		SecretKey:          secretKey,
	}

	data, err := msgpack.Marshal(cfg)
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set(configKey, string(data), nil)
		return err
	})
	return cfg, err
}
