package notifycfg

import (
	"encoding/json"
	"os"
	"sync"
)

// 定义 EmailSettings / DingTalkSettings 在这里或单独放到 model 包
type EmailSettings struct {
	SMTPServer string `json:"smtpServer"`
	Port       int    `json:"port"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	From       string `json:"from"`
}

type DingTalkSettings struct {
	Webhook string `json:"webhook"`
	Secret  string `json:"secret"`
}

var (
	emailSettings    *EmailSettings
	dingTalkSettings *DingTalkSettings
	mu               sync.RWMutex
)

func LoadEmailSettings() (*EmailSettings, error) {
	mu.RLock()
	defer mu.RUnlock()
	return emailSettings, nil
}

func SaveEmailSettings(cfg *EmailSettings) error {
	mu.Lock()
	defer mu.Unlock()
	emailSettings = cfg
	return saveToFile()
}

func LoadDingTalkSettings() (*DingTalkSettings, error) {
	mu.RLock()
	defer mu.RUnlock()
	return dingTalkSettings, nil
}

func SaveDingTalkSettings(cfg *DingTalkSettings) error {
	mu.Lock()
	defer mu.Unlock()
	dingTalkSettings = cfg
	return saveToFile()
}

// 简单保存到本地文件
func saveToFile() error {
	f, err := os.Create("notify_settings.json")
	if err != nil {
		return err
	}
	defer f.Close()
	data := map[string]interface{}{
		"email":    emailSettings,
		"dingtalk": dingTalkSettings,
	}
	return json.NewEncoder(f).Encode(data)
}
