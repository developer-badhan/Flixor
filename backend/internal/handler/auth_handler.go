package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"


	"github.com/developer-badhan/Flixor/internal/model"
	"github.com/developer-badhan/Flixor/internal/service"
	
)

/**
 * AuthHandler handles HTTP requests for authentication endpoints.
 * It owns exactly two jobs: parse the request, send the response.
 * All actual logic lives in AuthService — never in here.
*/
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler creates an AuthHandler wired to the given AuthService.
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Refresh token request struct
type refreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}
 
// Logout request struct
type logoutRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

// Register handles POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterRequest

	/**
	 *  ShouldBindJSON does two things at once:
	 * 1. Decodes the JSON body into req
	 * 2. Validates every field against the binding tags in model.RegisterRequest
	 *    (required, min=3, email format, etc.)
	 *  If either step fails, we return 400 immediately — service never called.
	*/
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": formatValidationError(err),
		})
		return
	}

	ip := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	// Delegate to service — no business logic here
	response, err := h.authService.Register(c.Request.Context(), &req, ip, userAgent)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err.Error() == "email already registered" {
			statusCode = http.StatusConflict // 409
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	// 201 Created — a new resource was created
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"message": "account created successfully",
		"data":    response,
	})
}

// Login handles POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": formatValidationError(err),
		})
		return
	}

	ip := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")

	response, err := h.authService.Login(c.Request.Context(), &req, ip, userAgent)
	if err != nil {
		// Invalid credentials → 401 Unauthorized
		statusCode := http.StatusUnauthorized
		if err.Error() == "login failed — please try again" {
			statusCode = http.StatusInternalServerError
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	// 200 OK — existing resource accessed successfully
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "login successful",
		"data":    response,
	})
}

// Refresh handles POST /api/v1/auth/refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	var req refreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "refresh_token is required",
		})
		return
	}
 
	ip := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
 
	tokens, err := h.authService.RefreshTokens(c.Request.Context(), req.RefreshToken, ip, userAgent)
	if err != nil {
		statusCode, message := mapAuthError(err)
		c.JSON(statusCode, gin.H{
			"success": false,
			"error":   message,
		})
		return
	}
 
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    tokens,
	})
}

// Logout handles POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	var req logoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "refresh_token is required",
		})
		return
	}
 
	if err := h.authService.Logout(c.Request.Context(), req.RefreshToken); err != nil {
		statusCode, message := mapAuthError(err)
		c.JSON(statusCode, gin.H{
			"success": false,
			"error":   message,
		})
		return
	}
 
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "logged out successfully",
	})
}

// LogoutAll handles POST /api/v1/auth/logout-all
func (h *AuthHandler) LogoutAll(c *gin.Context) {
	// The AuthMiddleware sets "userID" in the Gin context after validating the JWT.
	userIDStr, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "unauthorized",
		})
		return
	}
 
	userID, err := primitive.ObjectIDFromHex(userIDStr.(string))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid user ID in token",
		})
		return
	}
 
	if err := h.authService.LogoutAll(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "failed to revoke sessions",
		})
		return
	}
 
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "all sessions revoked successfully",
	})
}

/**
 * formatValidationError converts Gin's validation errors into a single
 * readable string. Without this, the raw error looks like:
 * "Key: 'RegisterRequest.Email' Error:Field validation for 'Email' failed on the 'email' tag"
 * With this, it becomes: "email must be a valid email address"
*/
func formatValidationError(err error) string {
	if err != nil {
		return err.Error()
	}
	return "invalid request"
}

/**
 * mapAuthError converts service-layer sentinel errors to HTTP status + message.
 * Centralizing this keeps all HTTP concerns in the handler layer.
*/
func mapAuthError(err error) (int, string) {
	switch {
	case errors.Is(err, service.ErrTokenNotFound):
		return http.StatusUnauthorized, "invalid refresh token"
	case errors.Is(err, service.ErrTokenBlacklisted):
		return http.StatusUnauthorized, "refresh token has been revoked"
	case errors.Is(err, service.ErrTokenExpired):
		return http.StatusUnauthorized, "refresh token has expired — please login again"
	case errors.Is(err, service.ErrTokenReuse):
		// Tell the client clearly: we detected potential theft and logged them out everywhere.
		return http.StatusUnauthorized, "security alert: token reuse detected — all sessions have been revoked"
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}
