package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func setupAdminRoutes(r *gin.Engine, db *gorm.DB) {
	adminAuth := r.Group("/admin")
	adminAuth.Use(adminAuthorisionMiddleware(db))
	adminAuth.GET("/", adminDashboard(db))

	adminAuth.POST("/create_user", adminCreatePOST(db))
	adminAuth.GET("/create_user", adminCreateGET())

	adminAuth.GET("/remove/:userid", adminRemoveUserGET(db))
}

func adminDashboard(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		u := c.Keys["user"].(user)

		var users []user
		if err := db.Find(&users).Error; err != nil {
			log.Println("Cant load users: ", err)
		}
		log.Println(u)
		c.Header("Cache-Control", "no-store")
		c.HTML(http.StatusOK, "admin_userlist.templ.html", gin.H{"Admin": u.Admin, "Users": users})
	}
}

func adminCreatePOST(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")
		isAdmin := c.PostForm("isAdmin") == "on"

		if len(username) == 0 || len(password) == 0 {
			c.Redirect(301, "/admin/create_user")
			return
		}

		if err := addUser(db, username, password, isAdmin); err != nil {
			log.Println(err)
			c.Redirect(301, "/admin/create_user")
			return
		}
		c.Redirect(http.StatusFound, "/admin")
	}
}

func adminCreateGET() gin.HandlerFunc {
	return func(c *gin.Context) {
		u := c.Keys["user"].(user)
		c.HTML(http.StatusOK, "admin_createuser.templ.html", gin.H{"Admin": u.Admin})
	}
}

func adminRemoveUserGET(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		userid := c.Param("userid")

		var currentUser user
		if err := db.Preload("Files").Where("guid = ?", userid).First(&currentUser).Error; err != nil {
			c.String(404, "Not found")
			return
		}

		for _, currentFile := range currentUser.Files {
			if os.Remove(currentFile.Path) != nil {
				c.String(500, "Could not remove")
				return
			}

			if err := db.Delete(&file{}, "guid = ?", currentFile.GUID).Error; err != nil {
				c.String(404, "Not found")
				return
			}
		}

		if err := db.Delete(&user{}, "guid = ?", userid).Error; err != nil {
			c.String(404, "Not found")
			return
		}

		c.Redirect(302, "/admin")
	}
}
