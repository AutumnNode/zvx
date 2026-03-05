package controller

import (
	"encoding/json"
	"fmt"
	"kube-api/kube"
	"kube-api/service"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

type TerminalMessage struct {
	Type string `json:"type"`
	Cols int    `json:"cols"`
	Rows int    `json:"rows"`
}

func PodShellWS(c *gin.Context) {
	ns := c.Query("namespace")
	pod := c.Query("pod")
	container := c.DefaultQuery("container", "")

	clientset, config := kube.InitClient()

	// 如果 container 为空，则获取 pod 的第一个容器
	if container == "" {
		podInfo, err := service.GetPod(ns, pod)
		if err != nil {
			fmt.Println("获取 pod 信息失败:", err)
			return
		}
		if len(podInfo.Spec.Containers) > 0 {
			container = podInfo.Spec.Containers[0].Name
		} else {
			fmt.Println("Pod 中没有找到容器")
			return
		}
	}

	// ✅ 使用全局 WsUpgrader，而不是自己定义
	conn, err := WsUpgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		fmt.Println("WebSocket 升级失败:", err)
		return
	}

	req := clientset.CoreV1().RESTClient().Post().
		Resource("pods").
		Name(pod).
		Namespace(ns).
		SubResource("exec").
		VersionedParams(&v1.PodExecOptions{
			Container: container,
			Command:   []string{"/bin/sh"},
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       true,
		}, scheme.ParameterCodec)

	executor, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		fmt.Println("创建 executor 出错:", err)
		_ = conn.Close()
		return
	}

	handler := &wsStreamHandler{conn: conn}
	err = executor.Stream(remotecommand.StreamOptions{
		Stdin:             handler,
		Stdout:            handler,
		Stderr:            handler,
		Tty:               true,
		TerminalSizeQueue: handler,
	})

	if err != nil {
		fmt.Println("执行出错:", err)
	}
	_ = conn.Close()
}

type wsStreamHandler struct {
	conn     *websocket.Conn
	sizeChan chan remotecommand.TerminalSize
}

func (w *wsStreamHandler) Read(p []byte) (int, error) {
	for {
		_, message, err := w.conn.ReadMessage()
		if err != nil {
			return 0, err
		}

		var msg TerminalMessage
		if err := json.Unmarshal(message, &msg); err == nil && msg.Type == "resize" {
			if w.sizeChan != nil {
				w.sizeChan <- remotecommand.TerminalSize{
					Width:  uint16(msg.Cols),
					Height: uint16(msg.Rows),
				}
			}
			continue
		}

		copy(p, message)
		return len(message), nil
	}
}

func (w *wsStreamHandler) Write(p []byte) (int, error) {
	return len(p), w.conn.WriteMessage(websocket.TextMessage, p)
}

func (w *wsStreamHandler) Next() *remotecommand.TerminalSize {
	if w.sizeChan == nil {
		w.sizeChan = make(chan remotecommand.TerminalSize, 1)
	}
	size := <-w.sizeChan
	return &size
}
