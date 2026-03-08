package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/mino/backend/internal/middleware"
	resp "github.com/mino/backend/internal/pkg/response"
	"github.com/mino/backend/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

type signInRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type changePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required,min=6"`
}

type changeUsernameRequest struct {
	NewUsername string `json:"new_username" binding:"required,min=2"`
	Password   string `json:"password" binding:"required"`
}

func (h *AuthHandler) SignIn(c *gin.Context) {
	var req signInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	tokens, err := h.authService.SignIn(req.Username, req.Password)
	if err != nil {
		resp.Unauthorized(c, err.Error())
		return
	}
	resp.OK(c, tokens)
}

func (h *AuthHandler) SignOut(c *gin.Context) {
	// With stateless JWT, sign-out is handled client-side
	// In a full implementation, we'd blacklist the token in Redis
	resp.OK(c, gin.H{"message": "signed out"})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	accessToken, err := h.authService.Refresh(req.RefreshToken)
	if err != nil {
		resp.Unauthorized(c, err.Error())
		return
	}
	resp.OK(c, gin.H{"access_token": accessToken})
}

func (h *AuthHandler) ChangePassword(c *gin.Context) {
	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.authService.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	resp.OK(c, gin.H{"message": "password changed"})
}

func (h *AuthHandler) ChangeUsername(c *gin.Context) {
	var req changeUsernameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}

	userID := middleware.GetUserID(c)
	if err := h.authService.ChangeUsername(userID, req.NewUsername, req.Password); err != nil {
		resp.BadRequest(c, err.Error())
		return
	}
	resp.OK(c, gin.H{"message": "username changed"})
}
