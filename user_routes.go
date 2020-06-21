package main

import (
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

func setupUserRoutes(r *gin.Engine, db *gorm.DB) {
	needAuth := r.Group("/user")
	needAuth.Use(userAuthorisionMiddleware(db))

	needAuth.GET("/", userGET(db))
	needAuth.GET("/file", fileUploadGET(db))
	needAuth.GET("/file/download/:fileid", downloadFileGET(db))
	needAuth.GET("/file/remove/:fileid", deleteFileGET(db))

	needAuth.POST("/file", filePOST(db))
}

func userGET(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		u := c.Keys["user"].(user)

		var files []file
		if err := db.Where("user_id = ?", u.Id).Find(&files).Error; err != nil {
			log.Println("Cant load user: ", err)
		}
		c.Header("Cache-Control", "no-store")
		c.HTML(http.StatusOK, "user_file_list.templ.html", gin.H{"Admin": u.Admin, "Files": files})
	}
}

func fileUploadGET(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		u := c.Keys["user"].(user)
		c.HTML(http.StatusOK, "user_fileupload.templ.html", gin.H{"Admin": u.Admin})
	}
}

func downloadFileGET(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		fileid := c.Param("fileid")

		var currentFile file
		if err := db.Where("guid = ?", fileid).First(&currentFile).Error; err != nil {
			c.String(404, "Not found")
			return
		}
		c.FileAttachment(currentFile.Path, currentFile.Name)
	}
}

func deleteFileGET(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		fileid := c.Param("fileid")

		var currentFile file
		if err := db.Where("guid = ?", fileid).First(&currentFile).Error; err != nil {
			c.String(404, "Not found")
			return
		}

		if os.Remove(currentFile.Path) != nil {
			c.String(500, "Could not remove")
			return
		}

		if err := db.Delete(&file{}, "guid = ?", fileid).Error; err != nil {
			c.String(404, "Not found")
			return
		}

		c.Redirect(302, "/user")
	}
}

func filePOST(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		u := c.Keys["user"].(user)

		uploadedFile, err := c.FormFile("file")
		if err != nil {
			c.String(http.StatusBadRequest, "Form issue")
			log.Println(err)
			return
		}

		obscuredName, err := GenerateHexToken(20)
		if err != nil {
			c.String(http.StatusBadRequest, "Bad")
			log.Println(err)
			return
		}

		if err := c.SaveUploadedFile(uploadedFile, "./uploads/"+obscuredName); err != nil {
			c.String(http.StatusBadRequest, "Uploading issue")
			log.Println(err)
			return
		}

		guid, err := GenerateHexToken(20)
		if err != nil {
			c.String(http.StatusBadRequest, "Bad")
			log.Println(err)
			return
		}

		newFile := &file{Name: uploadedFile.Filename, Path: "./uploads/" + obscuredName, UserId: u.Id, GUID: guid, ExpiresAt: time.Now().Add(1 * time.Hour).Unix()}

		if err := db.Save(newFile).Error; err != nil {
			c.String(http.StatusBadRequest, "Save")
			log.Println(err)
			return
		}

		c.Redirect(302, "/user")
	}
}
