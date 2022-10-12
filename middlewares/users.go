package middlewares

import (
	"net/http"

	"github.com/damascopaul/lfg-backend/endpoints"
	"github.com/damascopaul/lfg-backend/schemas"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	log "github.com/sirupsen/logrus"
)

// UserRequestBody adds the parsed request body to the context.
func UserRequestBody(c *gin.Context) {
	var req schemas.User
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
