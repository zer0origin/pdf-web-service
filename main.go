package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"log"
	"net/http"
	"pdf_service_web/controller"
)

type selectorInfo struct {
	documentUUID string
}

func main() {
	router := gin.Default()
	router.LoadHTMLGlob("templates/**/*.gohtml")
	router.Static("/css", "static/css")
	router.Static("/images", "static/images")
	router.Static("/js", "static/js")

	loginController := controller.LoginController{}
	router.GET("/login", loginController.LoginRender)
	router.POST("/login", loginController.LoginAuthHandler)

	router.GET("/selector", func(c *gin.Context) {
		if id, present := c.GetQuery("documentUUID"); present {
			_, err := uuid.Parse(id)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"Error": err.Error()})
				return
			}

			sel := selectorInfo{documentUUID: id}
			c.HTML(http.StatusOK, "selector", sel)
			return
		}

		c.JSON(http.StatusBadRequest, gin.H{"Error": "No param specified"})
	})

	log.Fatal(router.Run(":8080"))
}
