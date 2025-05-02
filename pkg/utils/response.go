package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一的响应结构
type Response struct {
	Code int         `json:"code"`    // 业务码
	Data interface{} `json:"data"`    // 数据
	Msg  string      `json:"message"` // 消息
}

// Pagination 分页参数
type Pagination struct {
	Current int   `json:"current"` // 当前页码
	Size    int   `json:"size"`    // 每页数量
	Total   int64 `json:"total"`   // 总记录数
}

// PageResult 分页结果
type PageResult struct {
	Records interface{} `json:"records"` // 记录列表
	Pagination
}

// 预定义业务状态码
const (
	SUCCESS          = 0
	ERROR            = -1
	UNAUTHORIZED     = 40100
	FORBIDDEN        = 40300
	NOT_FOUND        = 40400
	VALIDATION_ERROR = 40001
)

// 状态码对应的默认消息
var codeMessages = map[int]string{
	SUCCESS:          "操作成功",
	ERROR:            "操作失败",
	UNAUTHORIZED:     "未授权",
	FORBIDDEN:        "权限不足",
	NOT_FOUND:        "资源不存在",
	VALIDATION_ERROR: "参数验证失败",
}

// Success 成功响应
func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Code: SUCCESS,
		Data: data,
		Msg:  codeMessages[SUCCESS],
	})
}

// SuccessWithMessage 成功响应with自定义消息
func SuccessWithMessage(c *gin.Context, data interface{}, msg string) {
	c.JSON(http.StatusOK, Response{
		Code: SUCCESS,
		Data: data,
		Msg:  msg,
	})
}

// SuccessWithPage 分页成功响应
func SuccessWithPage(c *gin.Context, records interface{}, current, size int, total int64) {
	pageResult := PageResult{
		Records: records,
		Pagination: Pagination{
			Current: current,
			Size:    size,
			Total:   total,
		},
	}

	c.JSON(http.StatusOK, Response{
		Code: SUCCESS,
		Data: pageResult,
		Msg:  codeMessages[SUCCESS],
	})
}

// Error 错误响应
func Error(c *gin.Context, code int, msg string) {
	c.JSON(getHttpStatus(code), Response{
		Code: code,
		Data: nil,
		Msg:  msg,
	})
}

// ErrorWithData 错误响应with数据
func ErrorWithData(c *gin.Context, code int, msg string, data interface{}) {
	c.JSON(getHttpStatus(code), Response{
		Code: code,
		Data: data,
		Msg:  msg,
	})
}

// getHttpStatus 根据业务码获取对应的 HTTP 状态码
func getHttpStatus(code int) int {
	switch code {
	case UNAUTHORIZED:
		return http.StatusUnauthorized
	case FORBIDDEN:
		return http.StatusForbidden
	case NOT_FOUND:
		return http.StatusNotFound
	case VALIDATION_ERROR:
		return http.StatusBadRequest
	case SUCCESS:
		return http.StatusOK
	default:
		return http.StatusOK
	}
}
