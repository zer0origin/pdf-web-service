package main

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"log"
	"net/http"
	"pdf_service_web/controller"
	"pdf_service_web/keycloak"
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
	router.GET("/selector", selector)

	config := keycloak.RealmConfig{
		BaseUrl:      "http://localhost:8081",
		RealmName:    "pdf",
		Client:       "service-api",
		ClientSecret: "gtQLem8EJgxr537nbQlJh3Npd6Li6s0K",
	}
	loginController := controller.LoginController{
		AuthenticatedRedirect: "/user/",
		RealmConfig:           config,
	}
	router.GET("/", loginController.LoginRender)
	router.GET("/login", loginController.LoginRender)
	router.POST("/login", loginController.LoginAuthHandler)

	userController := controller.UserController{}
	router.GET("/user/", userController.UserInfo)
	router.GET("/user/info", userController.UserInfo)

	adminHandler, err := keycloak.NewAdminHandler(config)
	if err != nil {
		return
	}
	registerController := controller.RegistrationController{
		CreatedUserRedirect: "/",
		RealmConfig:         config,
		AdminHandler:        adminHandler,
	}
	router.GET("/register", registerController.RegisterRender)
	router.POST("/register", registerController.RegisterHandle)

	log.Fatal(router.Run(":8080"))
}

func selector(c *gin.Context) {
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
}
