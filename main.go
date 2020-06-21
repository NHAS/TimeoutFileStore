package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
)

const (
	CookieName = "token"
	TokenSize  = 128
)

type file struct {
	Id        int64
	CreatedAt time.Time
	ExpiresAt time.Time
	UserId    int64
	Path      string
	Name      string
	GUID      string `gorm:"unique;not null"`
}

type user struct {
	Id             int64     `form:"-"`
	GUID           string    `form:"-" gorm:"unique;not null"`
	Username       string    `form:"username" binding:"required" gorm:"unique;not null"`
	Password       string    `form:"password" binding:"required" gorm:"unique;not null"`
	Token          string    `form:"-"`
	TokenCreatedAt time.Time `form:"-"`
	Files          []file    `form:"-"`
	Admin          bool      `form:"-"`
}

func index(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		if valid, isAdmin, _, _ := checkCookie(c, db); valid {
			location := "/user"
			if isAdmin {
				location = "/admin"
			}

			c.Redirect(302, location)
			return
		}

		c.File("./resources/login.html")
	}
}

func main() {

	db, err := gorm.Open("sqlite3", "files.db")
	check(err)
	defer db.Close()

	db.AutoMigrate(&user{}, &file{})

	r := gin.Default()
	r.Static("/index_files", "./resources/index_files")
	//Probably a better way of loading these would be generating a slice using file walk
	r.LoadHTMLGlob("resources/*/*.templ.html")

	r.GET("/", index(db))

	setupSessionRoutes(r, db)
	setupAdminRoutes(r, db)
	setupUserRoutes(r, db)

	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Listen error: %s\n", err)
		}
	}()

	quit := make(chan os.Signal)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down....")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalln("Server was forced to shutdown: ", err)
	}
	log.Println("Done! Cya")

}
