package ws

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // TODO: ganti sesuai kebutuhan
	},
}

func ServeWS(w http.ResponseWriter, r *http.Request, userID string, role string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Println("WebSocket upgrade error:", err)
		return
	}

	client := &Client{
		UserID: userID,
		Role:   role,
		Conn:   conn,
		Send:   make(chan []byte, 256),
	}

	H.Register <- client

	go client.writePump()
	go client.readPump()
}

func (c *Client) readPump() {
	defer func() {
		H.Unregister <- c
		c.Conn.Close()
	}()

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}
		fmt.Println("📩 Received from", c.UserID, ":", string(message))
	}
}

func (c *Client) writePump() {
	defer func() {
		c.Conn.Close()
	}()

	for msg := range c.Send {
		err := c.Conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			break
		}
	}
}

func ServeWSHandler(c *gin.Context) {
	userID := c.Query("user_id")
	role := c.Query("role")

	if userID == "" || role == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id and role are required"})
		return
	}

	ServeWS(c.Writer, c.Request, userID, role)
}

// func ServeWSHandler(c *gin.Context) {
// 	userID := c.Query("user_id")
// 	role := c.Query("role")

// 	if userID == "" || role == "" {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id and role are required"})
// 		return
// 	}

// 	ServeWS(c.Writer, c.Request, userID, role)

// 	// ✅ Test kirim notifikasi ke semua admin setelah 3 detik
// 	go func() {
// 		time.Sleep(3 * time.Second)
// 		testNotif := map[string]interface{}{
// 			"title": "Test Notification",
// 			"body":  "Hello Admin, this is a test message",
// 		}
// 		data, _ := json.Marshal(testNotif)

// 		H.RoleSend <- RoleMessage{
// 			Role:    "admin",
// 			Message: data,
// 		}
// 		fmt.Println("📤 Sent test notification to admin role")
// 	}()
// }
