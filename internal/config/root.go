package config

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	aclmodel "github.com/jekiapp/nsqper/internal/model/acl"
	usergroup "github.com/jekiapp/nsqper/internal/repository/user"
	"github.com/tidwall/buntdb"
)

// CheckAndSetupRoot ensures the root group and root user exist in the DB, and sets them up if missing.
func CheckAndSetupRoot(db *buntdb.DB) error {
	// Check for root group
	rootGroup, _ := usergroup.GetGroupByName(db, "root")
	// Check for root user
	rootUser, _ := usergroup.GetUserByUsername(db, "root")

	if rootGroup != nil && rootUser != nil {
		return nil // Both exist
	}

	fmt.Println("Root group or root user not found. Setting up...")
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Set password for root user: ")
	password, _ := reader.ReadString('\n')
	password = strings.TrimSpace(password)
	if password == "" {
		return errors.New("password cannot be empty")
	}

	now := time.Now()

	// Create root group if missing
	if rootGroup == nil {
		rootGroup = &aclmodel.Group{
			ID:        uuid.NewString(),
			Name:      "root",
			CreatedAt: now,
			UpdatedAt: now,
		}
		err := usergroup.CreateGroup(db, *rootGroup)
		if err != nil {
			return errors.New("failed to create root group: " + err.Error())
		}
	}

	// Create root user if missing
	if rootUser == nil {
		hash := sha256.Sum256([]byte(password))
		hashedPassword := hex.EncodeToString(hash[:])
		rootUser = &aclmodel.User{
			ID:        uuid.NewString(),
			Username:  "root",
			Name:      "Root User",
			Password:  hashedPassword,
			Email:     "root@localhost",
			Type:      "admin",
			Status:    "active",
			CreatedAt: now,
			UpdatedAt: now,
		}
		err := usergroup.CreateUser(db, *rootUser)
		if err != nil {
			return errors.New("failed to create root user: " + err.Error())
		}
	}

	// Only create if not already assigned
	if _, err := usergroup.GetUserGroup(db, rootUser.ID, rootGroup.ID); err != nil {
		userGroup := aclmodel.UserGroup{
			ID:        uuid.NewString(),
			UserID:    rootUser.ID,
			GroupID:   rootGroup.ID,
			CreatedAt: now,
			UpdatedAt: now,
		}
		err := usergroup.CreateUserGroup(db, userGroup)
		if err != nil {
			return errors.New("failed to assign root user to root group: " + err.Error())
		}
	}

	fmt.Println("Root group and root user setup complete.")
	return nil
}
