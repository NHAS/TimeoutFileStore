package main

import (
	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func setupUserRoutes(r *gin.Engine, db *gorm.DB) {
	needAuth := r.Group("/user")
	needAuth.Use(userAuthorisionMiddleware(db))
	needAuth.GET("/", userGET(db))
	needAuth.GET("/file/:fileid", fileGET(db))
}

func userGET(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}

func fileGET(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {

	}
}
