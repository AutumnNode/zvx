package controller

import (
	"net/http"

	"github.com/gorilla/websocket"
)

// WebSocket 全局升级器，避免重复定义
var WsUpgrader = websocket.Upgrader{
	ReadBufferSize:  8192,
	WriteBufferSize: 8192,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}
