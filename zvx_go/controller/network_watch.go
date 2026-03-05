package controller

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

var networkUpgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// 模拟网络状态变化的数据结构
type NetworkUpdate struct {
	Event string `json:"event"` // 可为 "added"、"modified"、"deleted"
	ID    string `json:"id"`
	Name  string `json:"name"`
	Time  string `json:"time"`
}

// 处理 WebSocket 实时推送
func WatchNetwork(c *gin.Context) {
	namespace := c.Query("namespace")
	if namespace == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "namespace is required"})
		return
	}

	conn, err := networkUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket 升级失败:", err)
		return
	}
	defer conn.Close()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case t := <-ticker.C:
			// 模拟一个网络策略变更事件
			update := NetworkUpdate{
				Event: randomEvent(),
				ID:    randomID(),
				Name:  "policy-" + randomID()[0:4],
				Time:  t.Format(time.RFC3339),
			}
			msg, _ := json.Marshal(update)
			if err := conn.WriteMessage(websocket.TextMessage, msg); err != nil {
				log.Println("发送失败:", err)
				return
			}
		}
	}
}

func randomEvent() string {
	switch rand.Intn(3) {
	case 0:
		return "added"
	case 1:
		return "modified"
	default:
		return "deleted"
	}
}

func randomID() string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, 8)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
