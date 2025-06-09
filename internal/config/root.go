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
	aclmodel "github.com/jekiapp/topic-master/internal/model/acl"
	usergroup "github.com/jekiapp/topic-master/internal/repository/user"
	"github.com/tidwall/buntdb"
)

// CheckAndSetupRoot ensures the root group and root user exist in the DB, and sets them up if missing.
func CheckAndSetupRoot(db *buntdb.DB) error {
	// Check for root group
	rootFound, err := CheckRootGroupAndUserExist(db)
	if err != nil {
		return err
	}
	if rootFound {
		return nil
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

	rootGroup := aclmodel.Group{
		ID:          uuid.NewString(),
		Name:        aclmodel.GroupRoot,
		Description: "Main administrator group",
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	// fill it now, so that we can use it in the next step
	err = usergroup.CreateGroup(db, rootGroup)
	if err != nil {
		return errors.New("failed to create root group: " + err.Error())
	}

	hash := sha256.Sum256([]byte(password))
	hashedPassword := hex.EncodeToString(hash[:])
	rootUser := aclmodel.User{
		ID:        uuid.NewString(),
		Username:  aclmodel.GroupRoot,
		Name:      "Root User",
		Password:  hashedPassword,
		Status:    aclmodel.StatusUserActive,
		CreatedAt: now,
		UpdatedAt: now,
	}

	err = usergroup.CreateUser(db, rootUser)
	if err != nil {
		return errors.New("failed to create root user: " + err.Error())
	}

	// Only create if not already assigned
	userGroup := aclmodel.UserGroup{
		ID:        uuid.NewString(),
		UserID:    rootUser.ID,
		GroupID:   rootGroup.ID,
		Role:      aclmodel.RoleGroupAdmin,
		CreatedAt: now,
		UpdatedAt: now,
	}
	err = usergroup.CreateUserGroup(db, userGroup)
	if err != nil {
		return errors.New("failed to assign root user to root group: " + err.Error())
	}

	fmt.Println("Root group and root user setup complete.")
	return nil
}
