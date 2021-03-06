package main

import (
	"context"
	"encoding/json"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/csrf"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

const (
	CookieName = "token"
	TokenSize  = 128
)

type file struct {
	Id        int64
	CreatedAt int64
	ExpiresAt int64
	UserId    int64
	Path      string
	Name      string
	GUID      string `gorm:"unique;not null"`
}

type user struct {
	Id             int64  `form:"-"`
	GUID           string `form:"-" gorm:"unique;not null"`
	Username       string `form:"username" binding:"required" gorm:"unique;not null"`
	Password       string `form:"password" binding:"required" gorm:"unique;not null"`
	Token          string `form:"-"`
	TokenCreatedAt int64  `form:"-"`
	Files          []file `form:"-"`
	Admin          bool   `form:"-"`
}

func index(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if valid, isAdmin, _ := checkCookie(c, db); valid {
			location := "/user"
			if isAdmin {
				location = "/admin"
			}

			c.Redirect(302, location)
			return
		}
		c.Header("Cache-Control", "no-store")
		c.HTML(http.StatusOK, "login.templ.html", gin.H{csrf.TemplateTag: csrf.TemplateField(c.Request)})
	}
}

func fileExpiryChecker(db *gorm.DB, end chan bool) {

	for {
		select {
		case <-time.After(1 * time.Minute):
			var files []file
			if err := db.Find(&files).Error; err != nil {
				log.Println(err)
			}

			for _, f := range files {
				if time.Now().Unix() >= f.ExpiresAt {
					path := f.Path

					if err := db.Delete(&f).Error; err != nil {
						log.Println(err)
					}

					if err := os.Remove(path); err != nil {
						log.Println("Unable to remove ", path, " because: ", err)
					}

					log.Println("File removed: ", f.Path)

				}
			}

		case <-end:
			end <- true
			return

		}
	}

}

type config struct {
	CsrfKey                  string `json:"csrf_key"`
	Release                  bool   `json:"release_mode"`
	ListenAddress            string `json:"listen_addr"`
	DatabaseType             string `json:"database_type"`
	DatabaseConnectionString string `json:"database_connection"`
	UploadDirectory          string `json:"upload_directory"`
}

func main() {

	configBytes, err := ioutil.ReadFile("config.json")
	check(err)

	var c config
	err = json.Unmarshal(configBytes, &c)
	check(err)

	db, err := gorm.Open(c.DatabaseType, c.DatabaseConnectionString)
	check(err)
	defer db.Close()

	db.AutoMigrate(&user{}, &file{})

	r := gin.Default()

	if c.Release {
		gin.SetMode(gin.ReleaseMode)
	}
	r.SetFuncMap(template.FuncMap{
		"humanDate": humanDate,
	})

	r.Static("/index_files", "./resources/index_files")
	//Probably a better way of loading these would be generating a slice using file walk
	r.LoadHTMLGlob("resources/*/*.templ.html")

	CSRF := csrf.Protect([]byte(c.CsrfKey), csrf.Secure(c.Release))

	r.GET("/", index(db))

	setupSessionRoutes(r, db)
	setupAdminRoutes(r, db)
	setupUserRoutes(r, db, c.UploadDirectory)
	srv := &http.Server{
		Addr:    ":8080",
		Handler: CSRF(r),
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen error: %s\n", err)
		}
	}()

	exit := make(chan bool)

	go fileExpiryChecker(db, exit)

	quit := make(chan os.Signal)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	exit <- true
	log.Println("Shutting down....")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalln("Server was forced to shutdown: ", err)
	}
	<-exit
	log.Println("Done! Cya")

}
