package main

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"golang.org/x/crypto/bcrypt"
)

const (
	CookieName = "token"
)

type file struct {
	Id        int64
	CreatedAt time.Time
	ExpiresAt time.Time
	UserId    int64
	Path      string
	Name      string
	GUID      string
}

type user struct {
	Id             int64     `form:"-"`
	GUID           string    `form:"-"`
	Username       string    `form:"username" binding:"required"`
	Password       string    `form:"password" binding:"required"`
	Token          string    `form:"-"`
	TokenCreatedAt time.Time `form:"-"`
	Files          []file    `form:"-"`
	Admin          bool      `form:"-"`
}

func authenticatePOST(db *gorm.DB, dummyPassword []byte) gin.HandlerFunc {
	return func(c *gin.Context) {
		var u user
		if err := c.ShouldBindWith(&u, binding.Form); err != nil {
			log.Println(err)
			c.Redirect(302, "/")
			return
		}

		var record user
		if err := db.Where("username = ?", u.Username).First(&record).Error; err != nil {
			bcrypt.CompareHashAndPassword(dummyPassword, []byte(u.Password)) // Dummy compair to stop timing attacks
			c.Redirect(302, "/")
			log.Println(err)
			return
		}

		if err := bcrypt.CompareHashAndPassword([]byte(record.Password), []byte(u.Password)); err != nil {
			c.Redirect(302, "/")
			log.Println(err)
			return
		}

		u.Password = "" // Clear password from memory asap

		tokenBytes, err := GenerateRandomBytes(128)
		if err != nil {
			log.Println("Error generating random bytes for token: ", err)
			c.String(http.StatusInternalServerError, "Server error")
			return
		}

		token := hex.EncodeToString(tokenBytes)
		if db.Model(&record).Updates(user{Token: token, TokenCreatedAt: time.Now()}).Error != nil {
			log.Println("Error saving token in database: ", err)
			c.String(http.StatusInternalServerError, "Server error")
			return
		}

		c.SetCookie(CookieName, record.Username+":"+token, 3600, "", "localhost:8080", true, true)
		c.Redirect(302, "/user/"+record.GUID)
	}

}

func checkCookie(contents string, db *gorm.DB) bool {
	parts := strings.Split(contents, ":")
	if len(parts) != 2 {
		return false
	}

	var record user
	if db.Debug().Where("username = ? AND token = ?", parts[0], parts[1]).First(&record).Error != nil {
		return false
	}

	expiresAt := record.TokenCreatedAt.Add(1 * time.Hour)
	if time.Now().After(expiresAt) {
		return false
	}

	return true
}

func denyRequest(c *gin.Context) {
	c.Redirect(302, "/")
	c.Abort()
}

func userAuthorisionMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		contents, err := c.Cookie(CookieName)
		if err != nil {
			denyRequest(c)
			return
		}

		if !checkCookie(contents, db) {
			denyRequest(c)
			return
		}
	}
}

func userGET(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func fileGET(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func adminAuthorisionMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		contents, err := c.Cookie(CookieName)
		if err != nil {
			denyRequest(c)
			return
		}

		if !checkCookie(contents, db) {
			denyRequest(c)
			return
		}

		parts := strings.Split(contents, ":")
		if len(parts) != 2 { // This is checked in check cookie, but best be safe
			denyRequest(c)
			return
		}

		var record user
		if db.Debug().Where("username = ? AND token = ?", parts[0], parts[1]).First(&record).Error != nil {
			denyRequest(c)
			return
		}

		if !record.Admin {
			denyRequest(c)
			return
		}

	}
}

func adminDashboard(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func adminCreatePOST(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

	}
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

func check(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func main() {
	b, err := GenerateRandomBytes(16)
	check(err)

	dummyPassword, err := bcrypt.GenerateFromPassword(b, bcrypt.DefaultCost)
	check(err)

	db, err := gorm.Open("sqlite3", "files.db")
	check(err)
	defer db.Close()

	db.AutoMigrate(&user{}, &file{})

	r := gin.Default()
	r.Static("/index_files", "./resources/index_files")
	r.StaticFile("/", "./resources/login.html")
	r.POST("/authenticate", authenticatePOST(db, dummyPassword))

	needAuth := r.Group("/user")
	needAuth.Use(userAuthorisionMiddleware(db))
	needAuth.GET("/:userid", userGET(db))
	needAuth.GET("/:userid/file/:fileid", fileGET(db))

	adminAuth := r.Group("/admin")
	adminAuth.Use(adminAuthorisionMiddleware(db))
	adminAuth.GET("/", adminDashboard(db))
	adminAuth.POST("/create_user", adminCreatePOST(db))

	r.Run(":8080")
}
