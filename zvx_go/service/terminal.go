package service

import (
	"context"
	"fmt"
	"io"
	"net"
	"os/exec"
	"time"

	"github.com/creack/pty"
	"golang.org/x/crypto/ssh"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"kube-api/kube"
)

// ---------- 1. 获取控制节点 IP ----------
func GetControlNodeIP() (string, error) {
	clientset := kube.Clientset
	if clientset == nil {
		return "", fmt.Errorf("kube client not initialized")
	}

	nodes, err := clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{
		LabelSelector: "node-role.kubernetes.io/control-plane=true",
	})
	if err != nil {
		return "", err
	}

	if len(nodes.Items) == 0 {
		// 兼容旧版本 master 标签
		nodes, err = clientset.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{
			LabelSelector: "node-role.kubernetes.io/master=true",
		})
		if err != nil {
			return "", err
		}
	}

	if len(nodes.Items) == 0 {
		return "", fmt.Errorf("no control-plane node found")
	}

	// 取第一个控制节点 IP
	node := nodes.Items[0]
	for _, addr := range node.Status.Addresses {
		if addr.Type == v1.NodeInternalIP {
			return addr.Address, nil
		}
	}

	return "", fmt.Errorf("no internal IP found for control-plane node")
}

// ---------- 2. 认证 ----------
type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type AuthResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

func TerminalAuth(req AuthRequest, host string) AuthResponse {
	config := &ssh.ClientConfig{
		User: req.Username,
		Auth: []ssh.AuthMethod{
			ssh.Password(req.Password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
	}

	conn, err := ssh.Dial("tcp", net.JoinHostPort(host, "22"), config)
	if err != nil {
		return AuthResponse{Success: false, Message: "Invalid credentials or cannot connect"}
	}
	defer conn.Close()

	return AuthResponse{Success: true, Message: "Authentication successful"}
}

// ---------- 3. WebSocket + PTY ----------
func StartTerminal(wsConn io.ReadWriter) error {
	// 启动本地一个 bash 作为终端
	cmd := exec.Command("bash")

	// 创建伪终端
	ptmx, err := pty.Start(cmd)
	if err != nil {
		return err
	}
	defer func() {
		_ = ptmx.Close()
		_ = cmd.Process.Kill()
	}()

	// 双向绑定：WebSocket <-> PTY
	go func() {
		_, _ = io.Copy(ptmx, wsConn) // WebSocket 输入 -> PTY
	}()
	_, _ = io.Copy(wsConn, ptmx) // PTY 输出 -> WebSocket

	return nil
}
