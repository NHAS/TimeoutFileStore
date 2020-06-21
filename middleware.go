package main

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func checkCookie(c *gin.Context, db *gorm.DB) (valid, admin bool, userGUID, token string) {
	contents, err := c.Cookie(CookieName)
	if err != nil {
		return false, false, "", ""
	}

	parts := strings.Split(contents, ":")
	if len(parts) != 2 {
		return false, false, "", ""
	}

	var record user
	if db.Debug().Where("username = ? AND token = ?", parts[0], parts[1]).First(&record).Error != nil {
		return false, false, "", ""
	}

	expiresAt := record.TokenCreatedAt.Add(1 * time.Hour)
	if time.Now().After(expiresAt) {
		return false, false, "", ""
	}

	return true, record.Admin, record.GUID, parts[1]
}

func adminAuthorisionMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		valid, isAdmin, guid, _ := checkCookie(c, db)
		if !valid || !isAdmin {
			denyRequest(c)
			return
		}
		c.Keys = make(map[string]interface{}) // ??? I think this might be a bug in gin...

		c.Keys["guid"] = guid
	}
}

func userAuthorisionMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		valid, _, guid, _ := checkCookie(c, db)
		if !valid {
			denyRequest(c)
			return
		}
		c.Keys = make(map[string]interface{})
		c.Keys["guid"] = guid
	}
}
