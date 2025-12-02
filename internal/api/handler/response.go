package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// ListResponse 列表响应结构
type ListResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		Items any   `json:"items"`
		Total int64 `json:"total"`
		Page  int   `json:"page"`
		Size  int   `json:"size"`
	} `json:"data"`
}

// Success 成功响应
func Success(c *gin.Context, data any) {
	c.JSON(http.StatusOK, Response{
		Code:    0,
		Message: "success",
		Data:    data,
	})
}

// SuccessList 成功列表响应
func SuccessList(c *gin.Context, items any, total int64, page, size int) {
	resp := ListResponse{
		Code:    0,
		Message: "success",
	}
	resp.Data.Items = items
	resp.Data.Total = total
	resp.Data.Page = page
	resp.Data.Size = size
	c.JSON(http.StatusOK, resp)
}

// Error 错误响应
func Error(c *gin.Context, code int, message string) {
	c.JSON(code, Response{
		Code:    code,
		Message: message,
	})
}

// ErrorWithData 带数据的错误响应
func ErrorWithData(c *gin.Context, code int, message string, data any) {
	c.JSON(code, Response{
		Code:    code,
		Message: message,
		Data:    data,
	})
}
