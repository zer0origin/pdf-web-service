package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"pdf_service_web/controller"
	"pdf_service_web/controller/models"
	"pdf_service_web/jesr"
	"pdf_service_web/keycloak"
)

func main() {
	handleEmptyJesr := func(str string) { panic("Jesr Api Url must be present.") }
	mustNotBeEmpty(handleEmptyJesr, models.JESR_API_BASEURL)
	JesrApi := jesr.Api{BaseUrl: models.JESR_API_BASEURL}

	handleEmptyKeycloak := func(str string) { panic("Database login credentials must be present.") }
	mustNotBeEmpty(handleEmptyKeycloak, models.KEYCLOAK_BASEURL, models.KEYCLOAK_REALM_NAME, models.KEYCLOAK_CLIENT, models.KEYCLOAK_CLIENT_SECRET)
	config := &keycloak.RealmHandler{
		BaseUrl:      models.KEYCLOAK_BASEURL,
		RealmName:    models.KEYCLOAK_REALM_NAME,
		Client:       models.KEYCLOAK_CLIENT,
		ClientSecret: models.KEYCLOAK_CLIENT_SECRET,
	}

	router := gin.Default()
	router.LoadHTMLGlob("templates/**/*.gohtml")
	router.Static("/css", "static/css")
	router.Static("/images", "static/images")
	router.Static("/js", "static/js")

	adminHandler, err := keycloak.NewAdminHandler(config)
	if err != nil {
		panic(err.Error())
	}

	cloakSetup := &keycloak.Api{
		RealmHandler: config,
		AdminHandler: adminHandler,
	}

	middleware := controller.Middleware{
		Keycloak: cloakSetup,
	}

	loginController := controller.LoginController{
		AuthenticatedRedirect: "/user/",
		Keycloak:              cloakSetup,
		Middleware:            middleware,
	}

	router.GET("/", loginController.LoginRender)
	router.GET("/login", loginController.LoginRender)
	router.POST("/login", loginController.LoginAuthHandler)
	router.POST("/logout", loginController.Logout)

	userController := controller.UserController{
		KeycloakApi: cloakSetup,
		JesrApi:     JesrApi,
	}
	router.GET("/user/", middleware.RequireAuthenticated, userController.UserDashboard)
	router.GET("/user/info", middleware.RequireAuthenticated, userController.UserInfo)
	router.POST("/user/upload", middleware.RequireAuthenticated, userController.Upload)
	router.GET("/user/:uid/events", userController.PushNotifications)
	router.GET("/user/events/broadcast", userController.BroadcastNotification)

	registerController := controller.RegistrationController{
		CreatedUserRedirect: "/",
		KeycloakApi:         cloakSetup,
	}
	router.GET("/register", registerController.RegisterRender)
	router.POST("/register", registerController.RegisterHandle)

	log.Fatal(router.Run(":8080"))
}

func mustNotBeEmpty(errorHandle func(string), a ...string) {
	for _, s := range a {
		if len(s) == 0 {
			errorHandle(s)
		}
	}
}
