package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

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

	// Delegate to service — no business logic here
	response, err := h.authService.Register(c.Request.Context(), &req)
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

	response, err := h.authService.Login(c.Request.Context(), &req)
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
		"message": "login successful",
		"data":    response,
	})
}

/**
 * Me handles GET /api/v1/auth/me
 * Returns the currently authenticated user's profile.
 * Protected by auth middleware — user_id is guaranteed to be in context.
*/
func (h *AuthHandler) Me(c *gin.Context) {
	/**
	 * The middleware already validated the JWT and injected these values.
	 * If middleware is missing from the route, these will be empty strings —
	 * which is why every protected route must go through the middleware.
	*/
	userID := c.GetString("user_id")
	email := c.GetString("email")

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"user_id": userID,
			"email":   email,
		},
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