package endpoints

import (
	"net/http"
	"strings"

	"github.com/damascopaul/lfg-backend/schemas"

	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
)

// CloseGroup allows the user to mark a group as closed.
func CloseGroup(c *gin.Context) {
	g, _ := c.Keys["obj"].(schemas.Group)

	g.Status = -100 // Update the group status to closed.
	if err := g.Update(); err != nil {
		c.AbortWithStatusJSON(
			http.StatusInternalServerError, BodyInternalServerError)
		return
	}

	g.Password = "" // Makes sure the password is not included in the response.
	c.JSON(http.StatusOK, g)
	log.WithFields(
		log.Fields{"endpoint": "CloseGroup"}).Info("Request successful")
}

// CreateGroup creates a new group
func CreateGroup(c *gin.Context) {
	req, _ := c.Keys["req"].(schemas.Group)

	// Validate the request body
	if err := req.ValidateForCreate(); err != nil {
		// Return a 404 error if there are validation errors
		validationError, _ := err.(*schemas.ValidationError)
		c.AbortWithStatusJSON(http.StatusBadRequest, schemas.BodyError{
			Message:     err.Error(),
			FieldErrors: validationError.Errors,
		})
		return
	}

	if err := req.InitDB(); err != nil {
		c.AbortWithStatusJSON(
			http.StatusInternalServerError, BodyInternalServerError)
		return
	}

	req.OwnerID = c.GetInt64("user_id") // Set the ID of the user as owner.
	if err := req.Create(); err != nil {
		c.AbortWithStatusJSON(
			http.StatusInternalServerError, BodyInternalServerError)
		return
	}

	req.Password = ""
	c.JSON(http.StatusCreated, req)
	log.WithFields(
		log.Fields{"endpoint": "CreateGroup"}).Info("Request successful")
}

// JoinGroup allows a user to join a group
func JoinGroup(c *gin.Context) {
	g, _ := c.Keys["obj"].(schemas.Group)

	// Add the user as a member of the group.
	g.Members = append(g.Members, schemas.User{ID: c.GetInt64("user_id")})
	if err := g.Update(); err != nil {
		c.AbortWithStatusJSON(
			http.StatusInternalServerError, BodyInternalServerError)
		return
	}

	g.Password = "" // Makes sure the password is not included in the response.
	c.JSON(http.StatusOK, g)
	log.WithFields(log.Fields{"endpoint": "JoinGroup"}).Info("Request successful")
}

// KickFromGroup allows the owner to remove a member.
func KickFromGroup(c *gin.Context) {
	req, _ := c.Keys["req"].(schemas.User)
	g, _ := c.Keys["obj"].(schemas.Group)

	if err := req.InitDB(); err != nil {
		c.AbortWithStatusJSON(
			http.StatusInternalServerError, BodyInternalServerError)
		return
	}

	// Retrieve the user from the database.
	if err := req.Retrieve(); err != nil {
		if strings.Contains(err.Error(), "record not found") {
			// Return a 404 error if the group does not exist in the database
			c.AbortWithStatusJSON(http.StatusNotFound, BodyNotFound)
			return
		}
		// Return a 500 error for any other error other than "record not found"
		c.AbortWithStatusJSON(
			http.StatusInternalServerError, BodyInternalServerError)
		return
	}

	if !g.IsMember(req.ID) {
		// Return a 400 error if the user to kick is not a member of the group.
		log.WithFields(log.Fields{
			"details":  "The user to kick is not a member",
			"endpoint": "KickFromGroup",
			"group_id": g.ID,
			"user_id":  req.ID,
		}).Warning("Request failed")
		c.AbortWithStatusJSON(
			http.StatusBadRequest,
			schemas.BodyError{Message: "The user to kick is not a member"})
		return
	}

	if err := g.RemoveMember(req); err != nil {
		c.AbortWithStatusJSON(
			http.StatusInternalServerError, BodyInternalServerError)
		return
	}

	g.Password = "" // Makes sure the password is not included in the response.
	c.JSON(http.StatusOK, g)
	log.WithFields(
		log.Fields{"endpoint": "KickFromGroup"}).Info("Request successful")
}

// LeaveGroup allows a user to leave a group the user is a member of.
func LeaveGroup(c *gin.Context) {
	g, _ := c.Keys["obj"].(schemas.Group)
	u := schemas.User{ID: c.GetInt64("user_id")}

	if err := u.InitDB(); err != nil {
		c.AbortWithStatusJSON(
			http.StatusInternalServerError, BodyInternalServerError)
		return
	}

	if err := u.Retrieve(); err != nil {
		c.AbortWithStatusJSON(
			http.StatusInternalServerError, BodyInternalServerError)
		return
	}

	if err := g.RemoveMember(u); err != nil {
		c.AbortWithStatusJSON(
			http.StatusInternalServerError, BodyInternalServerError)
		return
	}

	g.Password = "" // Makes sure the password is not included in the response.
	c.JSON(http.StatusOK, g)
	log.WithFields(
		log.Fields{"endpoint": "LeaveGroup"}).Info("Request successful")
}

// ListGroups returns all the groups
func ListGroups(c *gin.Context) {
	g := schemas.Group{}

	if err := g.InitDB(); err != nil {
		c.AbortWithStatusJSON(
			http.StatusInternalServerError, BodyInternalServerError)
		return
	}

	groups, err := g.List()
	if err != nil {
		c.AbortWithStatusJSON(
			http.StatusInternalServerError, BodyInternalServerError)
		return
	}

	c.JSON(http.StatusOK, groups)
	log.WithFields(
		log.Fields{"endpoint": "ListGroups"}).Info("Request successful")
}

// RetrieveGroup returns the group details given its ID.
func RetrieveGroup(c *gin.Context) {
	// TODO: Add condition to show the group details if user is the owner
	// or password is correct.
	g, _ := c.Keys["obj"].(schemas.Group)

	g.Password = "" //Omits the password from the response
	c.JSON(http.StatusOK, g)
	log.WithFields(
		log.Fields{"endpoint": "RetrieveGroup"}).Info("Request successful")
}

// UpdateGroup allows the user to update the group details.
func UpdateGroup(c *gin.Context) {
	req, _ := c.Keys["req"].(schemas.Group)
	g, _ := c.Keys["obj"].(schemas.Group)

	// Checks for changes
	if req.Title != "" {
		g.Title = req.Title
	}
	if req.Description != "" {
		g.Description = req.Description
	}
	if req.MaxSize != 0 {
		g.MaxSize = req.MaxSize
	}

	if err := g.Update(); err != nil {
		c.AbortWithStatusJSON(
			http.StatusInternalServerError, BodyInternalServerError)
		return
	}

	g.Password = "" // Makes sure the password is not included in the response.
	c.JSON(http.StatusOK, g)
	log.WithFields(
		log.Fields{"endpoint": "UpdateGroup"}).Info("Request successful")
}

// UpdateGroupPassword allows the user to update the group details.
func UpdateGroupPassword(c *gin.Context) {
	req, _ := c.Keys["req"].(schemas.Group)
	g, _ := c.Keys["obj"].(schemas.Group)

	g.Password = req.Password // Set the new password
	if err := g.Update(); err != nil {
		c.AbortWithStatusJSON(
			http.StatusInternalServerError, BodyInternalServerError)
		return
	}

	g.Password = "" // Makes sure the password is not included in the response.
	log.WithFields(
		log.Fields{"endpoint": "UpdateGroupPassword"}).Info("Request successful")
	c.JSON(http.StatusOK, g)
}
