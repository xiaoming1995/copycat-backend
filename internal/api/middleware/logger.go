package middleware

import (
	"bytes"
	"io"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// LoggerMiddleware 自定义请求日志中间件
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()

		// 获取请求路径和方法
		path := c.Request.URL.Path
		method := c.Request.Method

		// 针对 POST/PUT 请求，记录 Body 内容（方便排查问题）
		var requestBody string
		if method == "POST" || method == "PUT" {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			// 将 Body 写回请求，否则后续 Handler 无法读取
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

			if len(bodyBytes) > 0 {
				// 截取前 500 个字符
				maxLen := 500
				if len(bodyBytes) < maxLen {
					maxLen = len(bodyBytes)
				}
				requestBody = string(bodyBytes[:maxLen])
			}
		}

		// 处理请求
		c.Next()

		// 结束时间
		endTime := time.Now()
		// 执行耗时
		latency := endTime.Sub(startTime)

		// 状态码
		statusCode := c.Writer.Status()
		// 客户端 IP
		clientIP := c.ClientIP()

		// 打印日志
		if requestBody != "" {
			log.Printf("[API-Log] %d | %13v | %15s | %-7s | %s | Body: %s",
				statusCode,
				latency,
				clientIP,
				method,
				path,
				requestBody,
			)
		} else {
			log.Printf("[API-Log] %d | %13v | %15s | %-7s | %s",
				statusCode,
				latency,
				clientIP,
				method,
				path,
			)
		}
	}
}
