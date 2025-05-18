package config

import (
	"errors"
	"os"

	"github.com/tidwall/buntdb"
	"github.com/vmihailenco/msgpack/v5"
)

type Config struct {
	NSQLookupdAddr string
	NSQDAddr       string
	SecretKeyFile  string
	SecretKey      []byte
}

func NewConfig(db *buntdb.DB) (*Config, error) {
	var cfg Config
	var raw string
	// Fetch the msgpack-encoded config from buntDB
	err := db.View(func(tx *buntdb.Tx) error {
		var err error
		raw, err = tx.Get("nsqper_config")
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

	// Read the secret key file
	secretKeyBytes, err := os.ReadFile(cfg.SecretKeyFile)
	if err != nil {
		return nil, errors.New("failed to read secret key file: " + err.Error())
	}
	cfg.SecretKey = secretKeyBytes

	return &cfg, nil
}
