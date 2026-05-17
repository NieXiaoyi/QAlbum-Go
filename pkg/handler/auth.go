package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"qalbum-server/pkg/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type LoginRequest struct {
	Code string `json:"code" binding:"required"`
}

type LoginResponse struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expires_in"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	token, expiresAt, err := h.authService.Login(req.Code)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	expiresIn := int(time.Until(expiresAt).Seconds())
	c.JSON(http.StatusOK, LoginResponse{
		Token:     token,
		ExpiresIn: expiresIn,
	})
}

func RegisterAuthRoutes(r *gin.Engine, authHandler *AuthHandler) {
	r.POST("/auth/login", authHandler.Login)
}
