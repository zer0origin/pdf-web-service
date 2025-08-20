package main

import (
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"pdf_service_web/controller"
	"pdf_service_web/controller/login"
	"pdf_service_web/controller/register"
	"pdf_service_web/controller/user"
	"pdf_service_web/jesr"
	"pdf_service_web/keycloak"
	"pdf_service_web/models"
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

	middleware := &controller.GinMiddleware{
		Keycloak: cloakSetup,
	}
	controller.SetMiddlewareInstance(middleware)

	loginController := &login.GinLogin{
		AuthenticatedRedirect: "/app",
		Keycloak:              cloakSetup,
		Middleware:            *middleware,
	}
	login.SetControllerInstance(loginController)

	router.GET("/", loginController.BaseRender)
	router.GET("/login", loginController.LoginRender)
	router.POST("/login", loginController.LoginAuthHandler)
	router.POST("/logout", loginController.Logout)

	userController := &user.GinUser{
		KeycloakApi: cloakSetup,
		JesrApi:     JesrApi,
	}
	user.SetControllerInstance(userController)

	router.GET("/app", middleware.RequireAuthenticated, userController.AppBase)
	router.GET("/user/details", middleware.RequireAuthenticated, userController.UserInfo)
	router.POST("/user/upload", BodySizeMiddleware(10*1024*1024), middleware.RequireAuthenticated, userController.Upload)
	router.GET("/user/dashboard", middleware.RequireAuthenticated, userController.UserDashboard)
	router.GET("/user/", middleware.RequireAuthenticated, userController.UserDashboard)
	router.GET("/user/events", userController.PushNotifications)
	router.POST("/user/events/broadcast", userController.BroadcastNotification)
	router.DELETE("/user/documents/:uid", middleware.RequireAuthenticated, userController.DeleteDocument)

	registerController := register.RegistrationController{
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

func BodySizeMiddleware(limit int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
		c.Next()
	}
}
