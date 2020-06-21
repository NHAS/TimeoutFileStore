package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func setupAdminRoutes(r *gin.Engine, db *gorm.DB) {
	adminAuth := r.Group("/admin")
	adminAuth.Use(adminAuthorisionMiddleware(db))
	adminAuth.GET("/", adminDashboard(db))
	adminAuth.POST("/create_user", adminCreatePOST(db))
	adminAuth.GET("/create_user", adminCreateGET(db))
}

func adminDashboard(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var users []user
		if err := db.Find(&users).Error; err != nil {
			log.Println("Cant load users: ", err)
		}

		c.HTML(http.StatusOK, "admin_dashboard.templ.html", gin.H{"Users": users})
	}
}

func adminCreatePOST(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func adminCreateGET(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}
