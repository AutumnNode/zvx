package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

var (
	infoLogger  *log.Logger
	errorLogger *log.Logger
)

// InitLogger 初始化日志系统
func InitLogger() {
	// 创建logs目录
	logDir := "./logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		log.Fatalf("无法创建logs目录: %v", err)
	}

	// 设置错误日志输出到文件
	errorLogFile := filepath.Join(logDir, "error.log")
	errorFile, err := os.OpenFile(errorLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("无法打开错误日志文件: %v", err)
	}
	errorLogger = log.New(errorFile, "", log.LstdFlags|log.Lshortfile)

	// 设置普通日志输出到文件和控制台
	infoLogFile := filepath.Join(logDir, "info.log")
	
	// 如果设置了DEBUG环境变量，同时输出到文件和控制台
	if os.Getenv("DEBUG") != "" {
		infoFile, err := os.OpenFile(infoLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			log.Fatalf("无法打开信息日志文件: %v", err)
		}
		multiWriter := io.MultiWriter(os.Stdout, infoFile)
		infoLogger = log.New(multiWriter, "", log.LstdFlags|log.Lshortfile)
	} else {
		// 只输出到控制台
		infoLogger = log.New(os.Stdout, "", log.LstdFlags)
	}

	// 记录启动信息
	LogInfo("日志系统初始化完成")
}

// LogInfo 记录普通信息日志
func LogInfo(format string, v ...interface{}) {
	if infoLogger != nil {
		prefix := fmt.Sprintf("[INFO] [%s] ", time.Now().Format("2006-01-02 15:04:05"))
		infoLogger.Printf(prefix+format, v...)
	}
}

// LogError 记录错误日志
func LogError(format string, v ...interface{}) {
	if errorLogger != nil {
		prefix := fmt.Sprintf("[ERROR] [%s] ", time.Now().Format("2006-01-02 15:04:05"))
		errorLogger.Printf(prefix+format, v...)
	}
	
	// 如果开启了DEBUG模式，同时输出到控制台
	if os.Getenv("DEBUG") != "" {
		if infoLogger != nil {
			prefix := fmt.Sprintf("[ERROR] [%s] [DEBUG] ", time.Now().Format("2006-01-02 15:04:05"))
			infoLogger.Printf(prefix+format, v...)
		}
	}
}

// LogPanic 记录panic信息
func LogPanic(v interface{}) {
	if errorLogger != nil {
		prefix := fmt.Sprintf("[PANIC] [%s] ", time.Now().Format("2006-01-02 15:04:05"))
		errorLogger.Printf(prefix+"%+v", v)
	}
	
	// 如果开启了DEBUG模式，同时输出到控制台
	if os.Getenv("DEBUG") != "" {
		if infoLogger != nil {
			prefix := fmt.Sprintf("[PANIC] [%s] [DEBUG] ", time.Now().Format("2006-01-02 15:04:05"))
			infoLogger.Printf(prefix+"%+v", v)
		}
	}
}
