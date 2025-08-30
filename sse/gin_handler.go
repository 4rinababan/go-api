package sse

import "github.com/gin-gonic/gin"

func GinHandler(c *gin.Context) {
	userID := c.Query("user_id") // contoh: ?user_id=UUID
	role := c.Query("role")      // contoh: ?role=admin
	ServeHTTP(c.Writer, c.Request, userID, role)
}
