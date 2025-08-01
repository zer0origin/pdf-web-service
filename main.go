package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type Page struct {
	Title          string
	Username       string
	HasMagicBall   bool
	HasCapturedRay bool
}

func main() {
	router := gin.Default()
	router.LoadHTMLGlob("templates/**/*.gohtml")
	router.Static("/css", "static/css")
	router.Static("/images", "static/images")
	router.Static("/js", "static/js")

	page := Page{
		Title:          "Selector",
		Username:       "BeamedCallum",
		HasMagicBall:   false,
		HasCapturedRay: true,
	}

	router.GET("/selector", func(c *gin.Context) {
		c.HTML(http.StatusOK, "selector", page)
	})

	log.Fatal(router.Run(":8080"))
}
