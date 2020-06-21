package main

import (
	"crypto/rand"
	"encoding/hex"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

func denyRequest(c *gin.Context) {
	c.Redirect(302, "/")
	c.Abort()
}

// GenerateRandomBytes returns securely generated random bytes.
// It will return an error if the system's secure random
// number generator fails to function correctly, in which
// case the caller should not continue.
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

func GenerateHexToken(n int) (string, error) {
	tokenBytes, err := GenerateRandomBytes(n)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(tokenBytes), nil
}

func addUser(db *gorm.DB, name, password string, admin bool) error {
	hashBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	guid, err := GenerateHexToken(16)
	if err != nil {
		return err
	}

	newUser := &user{Username: name, Password: string(hashBytes), GUID: guid, Admin: admin}

	return db.Debug().Create(newUser).Error
}

func check(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}
