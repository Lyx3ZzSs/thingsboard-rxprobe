package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/thingsboard-rxprobe/internal/model"
	"github.com/thingsboard-rxprobe/internal/service"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, "请求参数无效: "+err.Error())
		return
	}

	resp, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		Error(c, http.StatusUnauthorized, err.Error())
		return
	}

	Success(c, resp)
}

// CheckInit 检查系统是否已初始化
func (h *AuthHandler) CheckInit(c *gin.Context) {
	initialized, err := h.authService.CheckSystemInit(c.Request.Context())
	if err != nil {
		Error(c, http.StatusInternalServerError, "检查初始化状态失败")
		return
	}

	Success(c, gin.H{
		"initialized": initialized,
	})
}

// InitSystem 初始化系统
func (h *AuthHandler) InitSystem(c *gin.Context) {
	var req model.InitSystemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, "请求参数无效: "+err.Error())
		return
	}

	user, err := h.authService.InitSystem(c.Request.Context(), &req)
	if err != nil {
		Error(c, http.StatusBadRequest, err.Error())
		return
	}

	Success(c, user)
}

// GetCurrentUser 获取当前用户信息
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		Error(c, http.StatusUnauthorized, "未登录")
		return
	}

	user, err := h.authService.GetCurrentUser(c.Request.Context(), userID.(uint64))
	if err != nil {
		Error(c, http.StatusNotFound, "用户不存在")
		return
	}

	Success(c, user)
}

// ChangePassword 修改密码
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		Error(c, http.StatusUnauthorized, "未登录")
		return
	}

	var req model.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, "请求参数无效: "+err.Error())
		return
	}

	if err := h.authService.ChangePassword(c.Request.Context(), userID.(uint64), &req); err != nil {
		Error(c, http.StatusBadRequest, err.Error())
		return
	}

	Success(c, gin.H{
		"message": "密码修改成功",
	})
}

// Logout 登出（前端只需删除token即可，这里提供一个接口保持一致性）
func (h *AuthHandler) Logout(c *gin.Context) {
	Success(c, gin.H{
		"message": "登出成功",
	})
}
