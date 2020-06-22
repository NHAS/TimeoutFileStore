package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

func setupSessionRoutes(r *gin.Engine, db *gorm.DB) {

	b, err := GenerateRandomBytes(16)
	check(err)

	dummyPassword, err := bcrypt.GenerateFromPassword(b, bcrypt.DefaultCost)
	check(err)

	r.POST("/authenticate", authenticatePOST(db, dummyPassword))
	r.GET("/logout", logoutGET(db))
}

func authenticatePOST(db *gorm.DB, dummyPassword []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")

		if len(username) == 0 || len(password) == 0 {
			c.Redirect(302, "/")
			return
		}

		var record user
		if err := db.Where("username = ?", username).First(&record).Error; err != nil {
			bcrypt.CompareHashAndPassword(dummyPassword, []byte(password)) // Dummy compair to stop timing attacks
			c.Redirect(302, "/")
			log.Println(err)
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(record.Password), []byte(password)); err != nil {
			c.Redirect(302, "/")
			log.Println(err)
			return
		}

		password = "" // Clear password from memory asap

		token, err := GenerateHexToken(TokenSize)
		if err != nil {
			log.Println("Error generating token: ", err)
			c.String(http.StatusInternalServerError, "Server error")
			return
		}

		if db.Model(&record).Updates(user{Token: token, TokenCreatedAt: time.Now().Unix()}).Error != nil {
			log.Println("Error saving token in database: ", err)
			c.String(http.StatusInternalServerError, "Server error")
			return
		}

		c.SetSameSite(http.SameSiteStrictMode) // Stupid way of setting same site gin....
		c.SetCookie(CookieName, record.Username+":"+token, 3600, "", "localhost:8080", false, true)

		if record.Admin {
			c.Redirect(302, "/admin")
		} else {
			c.Redirect(302, "/user/")
		}

	}

}

func logoutGET(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		valid, _, u := checkCookie(c, db)
		if !valid {
			denyRequest(c)
			return
		}

		newToken, err := GenerateHexToken(TokenSize)
		if err != nil {
			log.Println("Error generating random bytes for token: ", err)
			c.String(http.StatusInternalServerError, "Server error")
			return
		}

		if err := db.Debug().Model(&user{}).Where("guid = ? AND token = ?", u.GUID, u.Token).Updates(user{Token: newToken, TokenCreatedAt: time.Now().Unix()}).Error; err != nil {
			log.Println("Error saving token in database: ", err)
			c.String(http.StatusInternalServerError, "Server error")
			return
		}

		c.Redirect(302, "/")

	}
}
