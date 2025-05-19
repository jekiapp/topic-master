package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/tidwall/buntdb"
	"github.com/vmihailenco/msgpack/v5"

	// Correct internal imports

	usergroup "github.com/jekiapp/nsqper/internal/repository/user"
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

func PromptConfig() *Config {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("NSQLookupdAddr: ")
	nsqlookupdAddr, _ := reader.ReadString('\n')
	// fmt.Print("NSQDAddr: ")
	// nsqdAddr, _ := reader.ReadString('\n')
	fmt.Print("Secret key file:")
	secretKeyFile, _ := reader.ReadString('\n')

	return &Config{
		NSQLookupdAddr: strings.TrimSpace(nsqlookupdAddr),
		// NSQDAddr:       strings.TrimSpace(nsqdAddr),
		SecretKeyFile: strings.TrimSpace(secretKeyFile),
	}
}

func SaveConfig(db *buntdb.DB, cfg *Config) error {
	data, err := msgpack.Marshal(cfg)
	if err != nil {
		return err
	}
	return db.Update(func(tx *buntdb.Tx) error {
		_, _, err := tx.Set("nsqper_config", string(data), nil)
		return err
	})
}

// CheckRootGroupAndUserExist returns true if both root group and root user exist in the DB.
func CheckRootGroupAndUserExist(db *buntdb.DB) (bool, error) {
	group, _ := usergroup.GetGroupByName(db, "root")
	user, _ := usergroup.GetUserByUsername(db, "root")
	return group != nil && user != nil, nil
}

//  TODO: add function to check if root group exists in db
// check also if root group  has "root" user

// if not, then create function to prompt to set up password for root user
// then create root group in db and user root with password, using user/user.go
