package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"golang.org/x/crypto/ssh"
)

// WebSocket 终端
func TerminalWS(c *gin.Context) {
	conn, err := WsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("WebSocket 升级失败:", err)
		return
	}
	defer conn.Close()

	// ✅ 从 terminal_api.go 获取认证信息
	auth := GetLastAuth()
	if auth.Username == "" || auth.Password == "" || auth.Host == "" {
		conn.WriteMessage(websocket.TextMessage, []byte("未认证，请先调用 /api/terminal-auth"))
		return
	}

	// 配置 SSH 客户端
	sshConfig := &ssh.ClientConfig{
		User:            auth.Username,
		Auth:            []ssh.AuthMethod{ssh.Password(auth.Password)},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	target := fmt.Sprintf("%s:%d", auth.Host, auth.Port)
	sshConn, err := ssh.Dial("tcp", target, sshConfig)
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("SSH连接失败: %v", err)))
		return
	}
	defer sshConn.Close()

	// 打开一个session
	session, err := sshConn.NewSession()
	if err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("创建SSH会话失败: %v", err)))
		return
	}
	defer session.Close()

	// 设置伪终端
	modes := ssh.TerminalModes{
		ssh.ECHO:          1,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	// 从查询参数获取终端大小
	colsStr := c.DefaultQuery("cols", "120")
	rowsStr := c.DefaultQuery("rows", "40")
	cols, err := strconv.Atoi(colsStr)
	if err != nil {
		cols = 120
	}
	rows, err := strconv.Atoi(rowsStr)
	if err != nil {
		rows = 40
	}

	if err := session.RequestPty("xterm", rows, cols, modes); err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("请求伪终端失败: %v", err)))
		return
	}

	// 启动 shell
	stdin, _ := session.StdinPipe()
	stdout, _ := session.StdoutPipe()
	stderr, _ := session.StderrPipe()

	if err := session.Shell(); err != nil {
		conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("启动shell失败: %v", err)))
		return
	}

	// ✅ SSH -> WebSocket
	writer := &wsWriter{conn: conn}
	// Combine stdout and stderr to avoid race conditions and interleaved writes
	multiReader := io.MultiReader(stdout, stderr)
	go func() {
		io.Copy(writer, multiReader)
	}()

	// ✅ WebSocket -> SSH
	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		// 尝试将消息解码为 resize 事件
		var resizeMessage struct {
			Type string `json:"type"`
			Cols int    `json:"cols"`
			Rows int    `json:"rows"`
		}

		if err := json.Unmarshal(msg, &resizeMessage); err == nil && resizeMessage.Type == "resize" {
			// 如果是 resize 事件, 更新 PTY 大小
			if session != nil {
				session.WindowChange(resizeMessage.Rows, resizeMessage.Cols)
			}
		} else {
			// 否则, 视为普通终端输入
			_, _ = stdin.Write(msg)
		}
	}
}

type wsWriter struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func (w *wsWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	err := w.conn.WriteMessage(websocket.TextMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}
