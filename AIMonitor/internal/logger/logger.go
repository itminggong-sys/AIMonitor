package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

// Config 日志配置
type Config struct {
	Level      string `yaml:"level" mapstructure:"level"`           // 日志级别
	Format     string `yaml:"format" mapstructure:"format"`         // 日志格式: json, text
	Output     string `yaml:"output" mapstructure:"output"`         // 输出方式: console, file, both
	FilePath   string `yaml:"file_path" mapstructure:"file_path"`   // 文件路径
	MaxSize    int    `yaml:"max_size" mapstructure:"max_size"`     // 单个日志文件最大大小（MB）
	MaxBackups int    `yaml:"max_backups" mapstructure:"max_backups"` // 保留的旧日志文件数量
	MaxAge     int    `yaml:"max_age" mapstructure:"max_age"`       // 保留日志文件的最大天数
	Compress   bool   `yaml:"compress" mapstructure:"compress"`     // 是否压缩旧日志文件
}

// DefaultConfig 默认配置
func DefaultConfig() Config {
	return Config{
		Level:      "info",
		Format:     "json",
		Output:     "console",
		FilePath:   "./logs/app.log",
		MaxSize:    100,
		MaxBackups: 5,
		MaxAge:     30,
		Compress:   true,
	}
}

// Init 初始化日志系统
func Init(cfg Config) error {
	// 设置日志级别
	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logrus.SetLevel(level)

	// 设置日志格式
	switch strings.ToLower(cfg.Format) {
	case "json":
		logrus.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	case "text":
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		})
	default:
		logrus.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: time.RFC3339,
		})
	}

	// 设置输出
	switch strings.ToLower(cfg.Output) {
	case "console":
		logrus.SetOutput(os.Stdout)
	case "file":
		if err := ensureLogDir(cfg.FilePath); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}
		file, err := os.OpenFile(cfg.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}
		logrus.SetOutput(file)
	case "both":
		if err := ensureLogDir(cfg.FilePath); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}
		file, err := os.OpenFile(cfg.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return fmt.Errorf("failed to open log file: %w", err)
		}
		// 创建多重写入器
		multiWriter := &MultiWriter{
			writers: []interface{}{
				os.Stdout,
				file,
			},
		}
		logrus.SetOutput(multiWriter)
	default:
		logrus.SetOutput(os.Stdout)
	}

	return nil
}

// ensureLogDir 确保日志目录存在
func ensureLogDir(filePath string) error {
	logDir := filepath.Dir(filePath)
	return os.MkdirAll(logDir, 0755)
}

// MultiWriter 多重写入器
type MultiWriter struct {
	writers []interface{}
}

// Write 实现io.Writer接口
func (mw *MultiWriter) Write(p []byte) (n int, err error) {
	for _, w := range mw.writers {
		if writer, ok := w.(interface{ Write([]byte) (int, error) }); ok {
			if _, err := writer.Write(p); err != nil {
				// 记录错误但继续写入其他writer
				continue
			}
		}
	}
	return len(p), nil
}

// GetLogger 获取带字段的日志器
func GetLogger(component string) *logrus.Entry {
	return logrus.WithField("component", component)
}

// WithFields 创建带字段的日志器
func WithFields(fields logrus.Fields) *logrus.Entry {
	return logrus.WithFields(fields)
}

// WithField 创建带单个字段的日志器
func WithField(key string, value interface{}) *logrus.Entry {
	return logrus.WithField(key, value)
}