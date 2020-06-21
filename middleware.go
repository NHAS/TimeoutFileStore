package main

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func checkCookie(c *gin.Context, db *gorm.DB) (valid, admin bool, u user) {
	contents, err := c.Cookie(CookieName)
	if err != nil {
		return false, false, u
	}

	parts := strings.Split(contents, ":")
	if len(parts) != 2 {
		return false, false, u
	}

	var record user
	if db.Debug().Where("username = ? AND token = ?", parts[0], parts[1]).First(&record).Error != nil {
		return false, false, u
	}

	expiresAt := time.Unix(record.TokenCreatedAt, 0).Add(1 * time.Hour)
	if time.Now().After(expiresAt) {
		return false, false, u
	}

	return true, record.Admin, record
}

func adminAuthorisionMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		valid, isAdmin, user := checkCookie(c, db)
		if !(valid && isAdmin) {
			denyRequest(c)
			return
		}
		c.Keys = make(map[string]interface{}) // ??? I think this might be a bug in gin...

		c.Keys["user"] = user
	}
}

func userAuthorisionMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		valid, _, user := checkCookie(c, db)
		if !valid {
			denyRequest(c)
			return
		}
		c.Keys = make(map[string]interface{})
		c.Keys["user"] = user
	}
}
