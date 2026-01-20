package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data,omitempty"`
}

// 状态码常量
const (
	CodeSuccess      = 0
	CodeBadRequest   = 400
	CodeUnauthorized = 401
	CodeForbidden    = 403
	CodeNotFound     = 404
	CodeServerError  = 500
)

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: CodeSuccess,
		Msg:  "success",
		Data: data,
	})
}

// SuccessWithMessage 成功响应带自定义消息
func SuccessWithMessage(c *gin.Context, msg string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: CodeSuccess,
		Msg:  msg,
		Data: data,
	})
}

// Error 错误响应
func Error(c *gin.Context, httpCode int, code int, msg string) {
	c.JSON(httpCode, Response{
		Code: code,
		Msg:  msg,
		Data: nil,
	})
}

// BadRequest 400 错误
func BadRequest(c *gin.Context, msg string) {
	Error(c, http.StatusBadRequest, CodeBadRequest, msg)
}

// Unauthorized 401 错误
func Unauthorized(c *gin.Context, msg string) {
	Error(c, http.StatusUnauthorized, CodeUnauthorized, msg)
}

// Forbidden 403 错误
func Forbidden(c *gin.Context, msg string) {
	Error(c, http.StatusForbidden, CodeForbidden, msg)
}

// NotFound 404 错误
func NotFound(c *gin.Context, msg string) {
	Error(c, http.StatusNotFound, CodeNotFound, msg)
}

// ServerError 500 错误
func ServerError(c *gin.Context, msg string) {
	Error(c, http.StatusInternalServerError, CodeServerError, msg)
}

// PageData 分页数据结构
type PageData struct {
	List     interface{} `json:"list"`
	Total    int64       `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

// SuccessWithPage 分页响应
func SuccessWithPage(c *gin.Context, list interface{}, total int64, page, pageSize int) {
	Success(c, PageData{
		List:     list,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}
