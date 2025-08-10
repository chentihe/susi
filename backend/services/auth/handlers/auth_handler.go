package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tihe/susi-auth-service/models"
	"github.com/tihe/susi-auth-service/services"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type AuthHandler struct {
	DB           *gorm.DB
	AdminService services.AdminService
}

func NewAuthHandler(db *gorm.DB, adminService services.AdminService) *AuthHandler {
	return &AuthHandler{
		DB:           db,
		AdminService: adminService,
	}
}

func generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func generateResetToken() (string, error) {
	b := make([]byte, 32)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		Email    string `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	secret, err := services.GenerateTOTPSecret(req.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate TOTP secret"})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	admin := models.Admin{
		Username:     req.Username,
		PasswordHash: string(hash),
		Email:        req.Email,
		TOTPSecret:   secret,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	if err := h.AdminService.CreateAdmin(&admin); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"totp_secret": secret})
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
		TOTP     string `json:"totp"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	var admin models.Admin
	if err := h.DB.Where("username = ?", req.Username).First(&admin).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid password"})
		return
	}
	if !services.ValidateTOTP(req.TOTP, admin.TOTPSecret) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid TOTP code"})
		return
	}
	token, err := services.GenerateJWT(admin.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token"})
		return
	}
	refreshToken, err := generateRefreshToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	if err := h.DB.Create(&models.RefreshToken{
		AdminID:   admin.ID,
		Token:     refreshToken,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store refresh token"})
		return
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
	})
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	cookie, err := c.Request.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "No refresh token"})
		return
	}
	var rt models.RefreshToken
	if err := h.DB.Where("token = ? AND expires_at > ?", cookie.Value, time.Now()).First(&rt).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired refresh token"})
		return
	}
	var admin models.Admin
	if err := h.DB.First(&admin, rt.AdminID).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
		return
	}
	// Rotate refresh token
	newRefreshToken, err := generateRefreshToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate refresh token"})
		return
	}
	expiresAt := time.Now().Add(7 * 24 * time.Hour)
	if err := h.DB.Model(&rt).Updates(models.RefreshToken{
		Token:     newRefreshToken,
		ExpiresAt: expiresAt,
		UpdatedAt: time.Now(),
	}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update refresh token"})
		return
	}
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "refresh_token",
		Value:    newRefreshToken,
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   true,
		Path:     "/",
	})
	token, err := services.GenerateJWT(admin.Username)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	cookie, err := c.Request.Cookie("refresh_token")
	if err == nil {
		h.DB.Where("token = ?", cookie.Value).Delete(&models.RefreshToken{})
		// Clear the cookie
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     "refresh_token",
			Value:    "",
			Expires:  time.Unix(0, 0),
			HttpOnly: true,
			Secure:   true,
			Path:     "/",
		})
	}
	c.JSON(http.StatusOK, gin.H{"message": "Logged out"})
}

func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	var req struct {
		Email string `json:"email"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	var admin models.Admin
	if err := h.DB.Where("email = ?", req.Email).First(&admin).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Email not found"})
		return
	}
	token, err := generateResetToken()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate reset token"})
		return
	}
	expiresAt := time.Now().Add(1 * time.Hour)
	resetToken := models.PasswordResetToken{
		AdminID:   admin.ID,
		Token:     token,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := h.DB.Create(&resetToken).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to store reset token"})
		return
	}
	// In a real app, email the token to the user
	c.JSON(http.StatusOK, gin.H{"reset_token": token})
}

func (h *AuthHandler) ResetPassword(c *gin.Context) {
	var req struct {
		Token       string `json:"token"`
		NewPassword string `json:"new_password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request"})
		return
	}
	var resetToken models.PasswordResetToken
	if err := h.DB.Where("token = ? AND expires_at > ?", req.Token, time.Now()).First(&resetToken).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid or expired reset token"})
		return
	}
	var admin models.Admin
	if err := h.DB.First(&admin, resetToken.AdminID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
		return
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}
	if err := h.DB.Model(&admin).Update("password_hash", string(hash)).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update password"})
		return
	}
	// Invalidate the reset token
	h.DB.Delete(&resetToken)
	c.JSON(http.StatusOK, gin.H{"message": "Password reset successful"})
}

func (h *AuthHandler) ValidateJWT(c *gin.Context) {
	token := c.GetHeader("Authorization")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing token"})
		return
	}
	claims, err := services.ValidateJWT(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"username": claims.Username, "exp": claims.ExpiresAt})
}

func RegisterAuthRoutes(rg *gin.RouterGroup, handler *AuthHandler) {
	rg.POST("/auth/register", handler.Register)
	rg.POST("/auth/login", handler.Login)
	rg.POST("/auth/refresh", handler.Refresh)
	rg.POST("/auth/logout", handler.Logout)
	rg.POST("/auth/forgot-password", handler.ForgotPassword)
	rg.POST("/auth/reset-password", handler.ResetPassword)
	rg.POST("/auth/validate", handler.ValidateJWT)
}
