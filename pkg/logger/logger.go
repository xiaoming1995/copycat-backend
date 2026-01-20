package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Level 日志级别
type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
)

// Logger 日志记录器
type Logger struct {
	level       Level
	debugLogger *log.Logger
	infoLogger  *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
	file        *os.File
}

// 全局日志实例
var defaultLogger *Logger

func init() {
	defaultLogger = New(os.Stdout, INFO)
}

// New 创建新的日志记录器
func New(out io.Writer, level Level) *Logger {
	flags := log.Ldate | log.Ltime | log.Lshortfile

	return &Logger{
		level:       level,
		debugLogger: log.New(out, "[DEBUG] ", flags),
		infoLogger:  log.New(out, "[INFO]  ", flags),
		warnLogger:  log.New(out, "[WARN]  ", flags),
		errorLogger: log.New(out, "[ERROR] ", flags),
	}
}

// InitFileLogger 初始化文件日志（同时输出到控制台和文件）
func InitFileLogger(logDir string, level Level) error {
	// 创建日志目录
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %w", err)
	}

	// 创建日志文件（按日期命名）
	logFileName := fmt.Sprintf("copycat_%s.log", time.Now().Format("2006-01-02"))
	logFilePath := filepath.Join(logDir, logFileName)

	file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %w", err)
	}

	// 同时输出到控制台和文件
	multiWriter := io.MultiWriter(os.Stdout, file)

	flags := log.Ldate | log.Ltime | log.Lshortfile

	defaultLogger = &Logger{
		level:       level,
		debugLogger: log.New(multiWriter, "[DEBUG] ", flags),
		infoLogger:  log.New(multiWriter, "[INFO]  ", flags),
		warnLogger:  log.New(multiWriter, "[WARN]  ", flags),
		errorLogger: log.New(multiWriter, "[ERROR] ", flags),
		file:        file,
	}

	// 同时设置标准 log 包的输出
	log.SetOutput(multiWriter)
	log.SetFlags(flags)

	log.Printf("日志文件已启用: %s", logFilePath)
	return nil
}

// GetLogFilePath 获取当前日志文件路径
func GetLogFilePath(logDir string) string {
	logFileName := fmt.Sprintf("copycat_%s.log", time.Now().Format("2006-01-02"))
	return filepath.Join(logDir, logFileName)
}

// Close 关闭日志文件
func Close() {
	if defaultLogger != nil && defaultLogger.file != nil {
		defaultLogger.file.Close()
	}
}

// SetLevel 设置日志级别
func SetLevel(level Level) {
	defaultLogger.level = level
}

// Debug 调试日志
func Debug(format string, v ...interface{}) {
	if defaultLogger.level <= DEBUG {
		defaultLogger.debugLogger.Printf(format, v...)
	}
}

// Info 信息日志
func Info(format string, v ...interface{}) {
	if defaultLogger.level <= INFO {
		defaultLogger.infoLogger.Printf(format, v...)
	}
}

// Warn 警告日志
func Warn(format string, v ...interface{}) {
	if defaultLogger.level <= WARN {
		defaultLogger.warnLogger.Printf(format, v...)
	}
}

// Error 错误日志
func Error(format string, v ...interface{}) {
	if defaultLogger.level <= ERROR {
		defaultLogger.errorLogger.Printf(format, v...)
	}
}

// Fatal 致命错误日志 (会导致程序退出)
func Fatal(format string, v ...interface{}) {
	defaultLogger.errorLogger.Fatalf(format, v...)
}
