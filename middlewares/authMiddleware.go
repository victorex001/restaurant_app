package middleware

import (
	"net/http"
	"restaurant_app/helpers"

	"github.com/gin-gonic/gin"
)

func Authentication() gin.HandlerFunc{
	return func(c *gin.Context){
		clientToken := c.Request.Header.Get("token")

		if clientToken == ""{
			c.JSON(http.StatusInternalServerError, gin.H{"error": "No Authourization header provided"})
			c.Abort()
			return
		}
		claims, err := helpers.ValidateToken(clientToken)
		if err != ""{
			c.JSON(http.StatusInternalServerError, gin.H{"error": err})
			c.Abort()
			return
		}
		c.Set("email", claims.Email)
		c.Set("firstName", claims.FirstName)
		c.Set("lastName", claims.LastName)
		c.Set("uid", claims.Uid)

		c.Next()
	}
}