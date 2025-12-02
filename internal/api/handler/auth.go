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
		Error(c, http.StatusBadRequest, err.Error())
		return
	}

	resp, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		Error(c, http.StatusUnauthorized, err.Error())
		return
	}

	Success(c, resp)
}

// Logout 用户登出
func (h *AuthHandler) Logout(c *gin.Context) {
	// 前端清除 token 即可，后端无需处理
	Success(c, nil)
}

// GetCurrentUser 获取当前用户信息
func (h *AuthHandler) GetCurrentUser(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		Error(c, http.StatusUnauthorized, "未登录")
		return
	}

	userClaims := claims.(*service.Claims)
	Success(c, gin.H{
		"user_id":  userClaims.UserID,
		"username": userClaims.Username,
		"role":     userClaims.Role,
	})
}

// ChangePassword 修改密码
func (h *AuthHandler) ChangePassword(c *gin.Context) {
	claims, exists := c.Get("claims")
	if !exists {
		Error(c, http.StatusUnauthorized, "未登录")
		return
	}

	var req model.ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		Error(c, http.StatusBadRequest, err.Error())
		return
	}

	userClaims := claims.(*service.Claims)
	if err := h.authService.ChangePassword(c.Request.Context(), userClaims.UserID, &req); err != nil {
		Error(c, http.StatusBadRequest, err.Error())
		return
	}

	Success(c, nil)
}
