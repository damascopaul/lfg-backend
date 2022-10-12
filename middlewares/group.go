package middlewares

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/damascopaul/lfg-backend/endpoints"
	"github.com/damascopaul/lfg-backend/schemas"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	log "github.com/sirupsen/logrus"
)

// GroupObject adds the Group entry to the context.
func GroupObject(c *gin.Context) {
	// Parse the group ID from the URL parameter.
	gid, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		// Return a 500 error if there was an error in parsing
		// the group ID in the URL
		log.Errorf("Could not parse ID parameter from URL. Error: %v", err)
		c.AbortWithStatusJSON(
			http.StatusInternalServerError, endpoints.BodyInternalServerError)
		return
	}

	g := schemas.Group{}
	if err := g.InitDB(); err != nil {
		// Return a 500 error for any other error other than "record not found"
		c.AbortWithStatusJSON(
			http.StatusInternalServerError, endpoints.BodyInternalServerError)
		return
	}

	g.ID = gid
	if err := g.RetrieveWithPassword(); err != nil {
		if strings.Contains(err.Error(), "record not found") {
			// Return a 404 error if the group does not exist in the database
			c.AbortWithStatusJSON(http.StatusNotFound, endpoints.BodyNotFound)
			return
		}
		// Return a 500 error for any other error other than "record not found"
		c.AbortWithStatusJSON(
			http.StatusInternalServerError, endpoints.BodyInternalServerError)
		return
	}

	c.Set("obj", g)
	c.Next()
}

// GroupRequestBody adds the request body to the context.
func GroupRequestBody(c *gin.Context) {
	var req schemas.Group
	if err := c.ShouldBindWith(&req, binding.JSON); err != nil {
		log.WithFields(log.Fields{
			"error": err.Error(),
		}).Error("Failed to bind JSON request body")
		c.AbortWithStatusJSON(
			http.StatusInternalServerError, endpoints.BodyInternalServerError)
		return
	}

	c.Set("req", req)
	c.Next()
}

// AllowIfGroupIsNotFull allows requests for groups that are not yet full.
func AllowIfGroupIsNotFull(c *gin.Context) {
	g, ok := c.Keys["obj"].(schemas.Group)
	if !ok {
		c.AbortWithStatusJSON(
			http.StatusInternalServerError, endpoints.BodyInternalServerError)
		return
	}

	if g.IsFull() {
		// Return a 400 error if the group is full
		log.WithFields(log.Fields{
			"permission": "AllowIfGroupIsNotFull",
			"details":    "Request denied because the group is full",
			"group_id":   g.ID,
		}).Info("Permission error")
		c.AbortWithStatusJSON(
			http.StatusBadRequest,
			schemas.BodyError{Message: "Group is full"})
	}

	c.Next()
}

// AllowIfUserIsNotMember allows requests on groups where the user is not a member.
func AllowIfUserIsNotMember(c *gin.Context) {
	g, ok := c.Keys["obj"].(schemas.Group)
	if !ok {
		c.AbortWithStatusJSON(
			http.StatusInternalServerError, endpoints.BodyInternalServerError)
		return
	}

	uid := c.GetInt64("user_id")
	if g.IsMember(uid) {
		// Return a 400 error if the user is a member of the group
		log.WithFields(log.Fields{
			"permission": "AllowIfUserIsNotMember",
			"details":    "Request denied because the user is a member of the group",
			"group_id":   g.ID,
			"user_id":    uid,
		}).Info("Permission error")
		c.AbortWithStatusJSON(
			http.StatusBadRequest,
			schemas.BodyError{Message: "User is a member of the group"})
		return
	}

	c.Next()
}

// AllowIfUserIsNotOwner allows requests on groups where the user is not the owner.
func AllowIfUserIsNotOwner(c *gin.Context) {
	g, ok := c.Keys["obj"].(schemas.Group)
	if !ok {
		c.AbortWithStatusJSON(
			http.StatusInternalServerError, endpoints.BodyInternalServerError)
		return
	}

	uid := c.GetInt64("user_id")
	if g.IsOwner(uid) {
		// Return a 400 error if the user is the owner of the group.
		log.WithFields(log.Fields{
			"permission": "AllowIfUserIsNotOwner",
			"details":    "Request denied because the user is the owner of the group",
			"group_id":   g.ID,
			"user_id":    uid,
		}).Info("Permission error")
		c.AbortWithStatusJSON(
			http.StatusBadRequest,
			schemas.BodyError{Message: "User is the owner of the group"})
		return
	}

	c.Next()
}

// AllowIfUserIsOwner allows requests on groups where the user is the owner.
func AllowIfUserIsOwner(c *gin.Context) {
	g, ok := c.Keys["obj"].(schemas.Group)
	if !ok {
		c.AbortWithStatusJSON(
			http.StatusInternalServerError, endpoints.BodyInternalServerError)
		return
	}

	uid := c.GetInt64("user_id")
	if !g.IsOwner(uid) {
		// Return a 400 error if the user is the owner of the group.
		log.WithFields(log.Fields{
			"permission": "AllowIfUserIsOwner",
			"details":    "Request denied because the user is not the owner of the group",
			"group_id":   g.ID,
			"user_id":    uid,
		}).Info("Permission error")
		c.AbortWithStatusJSON(
			http.StatusBadRequest,
			schemas.BodyError{Message: "User is not the owner of the group"})
		return
	}

	c.Next()
}

// AllowIfUserIsMember allows requests on groups where the user is a member.
func AllowIfUserIsMember(c *gin.Context) {
	g, ok := c.Keys["obj"].(schemas.Group)
	if !ok {
		c.AbortWithStatusJSON(
			http.StatusInternalServerError, endpoints.BodyInternalServerError)
		return
	}

	uid := c.GetInt64("user_id")
	if !g.IsMember(uid) {
		// Return a 400 error if the user is not a member of the group
		log.WithFields(log.Fields{
			"permission": "AllowIfUserIsMember",
			"details":    "Request denied because the user is not a member of the group",
			"group_id":   g.ID,
			"user_id":    uid,
		}).Info("Permission error")
		c.AbortWithStatusJSON(
			http.StatusBadRequest,
			schemas.BodyError{Message: "User is not a member of the group"})
		return
	}

	c.Next()
}

// AllowIfCorrectGroupPassword allows requests if the group password is correct.
func AllowIfCorrectGroupPassword(c *gin.Context) {
	g, ok := c.Keys["obj"].(schemas.Group)
	if !ok {
		c.AbortWithStatusJSON(http.StatusNotFound, endpoints.BodyNotFound)
		return
	}

	// No need to check if the group is not private.
	if !g.IsPrivate() {
		c.Next()
		return
	}

	// Check if the user has the correct group password
	var req schemas.Group
	if err := c.ShouldBindWith(&req, binding.JSON); err != nil {
		log.WithFields(log.Fields{
			"details": "Failed to bind JSON in AllowIfCorrectGroupPassword",
			"error":   err.Error(),
		}).Error("Failed to bind JSON request body")
		if err.Error() == "EOF" {
			// Return a 400 error if there is no request body.
			c.AbortWithStatusJSON(
				http.StatusBadRequest,
				schemas.BodyError{Message: "Group password is required"})
			return
		}
		// Return a 500 error for errors other than the EOF error.
		c.AbortWithStatusJSON(
			http.StatusInternalServerError, endpoints.BodyInternalServerError)
		return
	}
	if err := g.ValidatePassword(req.Password); err != nil {
		// Return a 403 error if the group password does not match
		// the one on the request body.
		c.AbortWithStatusJSON(
			http.StatusForbidden, schemas.BodyError{Message: "Incorrect password"})
		return
	}

	c.Next()
}

// AllowIfGroupIsOpen allows requests if the group is open.
func AllowIfGroupIsOpen(c *gin.Context) {
	g, ok := c.Keys["obj"].(schemas.Group)
	if !ok {
		c.AbortWithStatusJSON(
			http.StatusInternalServerError, endpoints.BodyInternalServerError)
		return
	}

	if !g.IsOpen() {
		// Return a 400 error if the group is not open.
		log.WithFields(log.Fields{
			"permission": "AllowIfUserIsMember",
			"details":    "Request denied because the group is not open",
			"group_id":   g.ID,
		}).Info("Permission error")
		c.AbortWithStatusJSON(
			http.StatusBadRequest,
			schemas.BodyError{Message: "Group is not open"})
		return
	}

	c.Next()
}
