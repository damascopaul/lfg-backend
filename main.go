package main

import (
	"github.com/damascopaul/lfg-backend/endpoints"
	"github.com/damascopaul/lfg-backend/middlewares"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

func GetAPI() *gin.Engine {
	api := gin.Default()

	// Routes
	privateEndpoints := api.Group("/")
	privateEndpoints.Use(middlewares.AuthenticateRequests)
	{
		privateEndpoints.POST(
			"/groups/:id/close", middlewares.GroupObject,
			middlewares.AllowIfUserIsOwner, middlewares.AllowIfGroupIsOpen,
			endpoints.CloseGroup)
		privateEndpoints.GET("/groups", endpoints.ListGroups)
		privateEndpoints.POST(
			"/groups", middlewares.GroupRequestBody, endpoints.CreateGroup)
		privateEndpoints.PATCH(
			"groups/:id", middlewares.GroupObject, middlewares.AllowIfUserIsOwner,
			middlewares.AllowIfGroupIsOpen, middlewares.GroupRequestBody,
			endpoints.UpdateGroup)
		privateEndpoints.PATCH(
			"groups/:id/password", middlewares.GroupObject,
			middlewares.AllowIfUserIsOwner, middlewares.AllowIfGroupIsOpen,
			middlewares.GroupRequestBody, endpoints.UpdateGroupPassword)
		privateEndpoints.GET(
			"/groups/:id", middlewares.GroupObject, endpoints.RetrieveGroup)
		privateEndpoints.POST(
			"/groups/:id/join", middlewares.GroupObject,
			middlewares.AllowIfGroupIsNotFull, middlewares.AllowIfUserIsNotMember,
			middlewares.AllowIfUserIsNotOwner, middlewares.AllowIfGroupIsOpen,
			middlewares.AllowIfCorrectGroupPassword,
			endpoints.JoinGroup)
		privateEndpoints.POST(
			"/groups/:id/leave", middlewares.GroupObject,
			middlewares.AllowIfGroupIsOpen, middlewares.AllowIfUserIsMember,
			endpoints.LeaveGroup)
		privateEndpoints.POST(
			"groups/:id/kick", middlewares.UserRequestBody, middlewares.GroupObject,
			middlewares.AllowIfGroupIsOpen, middlewares.AllowIfUserIsOwner,
			endpoints.KickFromGroup)
	}
	api.POST("/sign-up", middlewares.UserRequestBody, endpoints.SignUp)
	api.POST("/sign-in", middlewares.UserRequestBody, endpoints.SignIn)
	return api
}

func main() {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.DebugLevel) // TODO: Should be conditional based on env.
	api := GetAPI()
	api.Run("localhost:8080")
}
