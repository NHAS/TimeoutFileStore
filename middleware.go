package main

import (
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func checkCookie(c *gin.Context, db *gorm.DB) (valid, admin bool, username, token string) {
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

	return true, record.Admin, parts[0], parts[1]
}

func adminAuthorisionMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if valid, isAdmin, _, _ := checkCookie(c, db); !valid || !isAdmin {
			denyRequest(c)
			return
		}
	}
}

func userAuthorisionMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if valid, _, _, _ := checkCookie(c, db); valid {
			denyRequest(c)
			return
		}
	}
}
