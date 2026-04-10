package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/developer-badhan/Flixor/internal/model"
	"github.com/developer-badhan/Flixor/internal/repository"
	"github.com/developer-badhan/Flixor/internal/service"
)

/**
 * UserHandler handles all user-profile HTTP endpoints.
 * Parse request → call service → send response. No business logic here.
*/
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler wires the handler to UserService.
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

/**
 * Me handles GET /api/v1/user/me
 * Returns the full public profile of the authenticated user.
 * "userID" is guaranteed to be in context because this route sits behind Auth middleware.
*/
func (h *UserHandler) Me(c *gin.Context) {
	userID := c.GetString("userID") // matches c.Set("userID", ...) in middleware

	profile, err := h.userService.GetMe(c.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "failed to retrieve profile"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "data": profile})
}

/**
 * UpdateProfile handles PATCH /api/v1/user/profile
 * Accepts JSON body with optional `username` and `password` fields.
 * Email is never accepted here — that restriction is enforced in the service.
*/
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	var req model.UpdateProfileRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	userID := c.GetString("userID")

	updated, err := h.userService.UpdateProfile(c.Request.Context(), userID, &req)
	if err != nil {
		statusCode, msg := mapUserError(err)
		c.JSON(statusCode, gin.H{"success": false, "error": msg})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "profile updated",
		"data":    updated,
	})
}

/**
 * UploadProfilePicture handles POST /api/v1/user/profile-picture
 * Accepts multipart/form-data with a single file field named "picture".
 * Max 5 MB. Allowed types: jpeg, png, webp.
*/
func (h *UserHandler) UploadProfilePicture(c *gin.Context) {
	userID := c.GetString("userID")

	fileHeader, err := c.FormFile("picture")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "field 'picture' is required (multipart/form-data)",
		})
		return
	}

	updated, err := h.userService.UploadProfilePicture(c.Request.Context(), userID, fileHeader)
	if err != nil {
		statusCode, msg := mapUserError(err)
		c.JSON(statusCode, gin.H{"success": false, "error": msg})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "profile picture updated",
		"data":    updated,
	})
}

/**
 * SendOTP handles POST /api/v1/user/send-otp
 * Generates an OTP and sends it to the authenticated user's email.
 * Idempotent — calling it again simply overwrites the previous OTP and resets the TTL.
*/
func (h *UserHandler) SendOTP(c *gin.Context) {
	userID := c.GetString("userID")

	if err := h.userService.SendOTP(c.Request.Context(), userID); err != nil {
		statusCode, msg := mapUserError(err)
		c.JSON(statusCode, gin.H{"success": false, "error": msg})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "verification code sent to your email",
	})
}

/**
 * VerifyOTP handles POST /api/v1/user/verify-otp
 * Body: {"otp": "123456"}
*/
func (h *UserHandler) VerifyOTP(c *gin.Context) {
	var req model.VerifyOTPRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "otp is required and must be exactly 6 digits",
		})
		return
	}

	userID := c.GetString("userID")

	if err := h.userService.VerifyOTP(c.Request.Context(), userID, req.OTP); err != nil {
		statusCode, msg := mapUserError(err)
		c.JSON(statusCode, gin.H{"success": false, "error": msg})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "email verified successfully",
	})
}

/**
 * mapUserError converts UserService sentinel errors to HTTP status + message.
 * All HTTP concerns live in the handler layer — the service returns typed errors only.
*/
func mapUserError(err error) (int, string) {
	switch {
	case errors.Is(err, service.ErrAlreadyVerified):
		return http.StatusConflict, "account is already verified"
	case errors.Is(err, service.ErrOTPNotFound):
		return http.StatusBadRequest, "no OTP pending — request a new verification code"
	case errors.Is(err, service.ErrOTPExpired):
		return http.StatusUnauthorized, "OTP has expired — request a new verification code"
	case errors.Is(err, service.ErrOTPInvalid):
		return http.StatusUnauthorized, "invalid OTP"
	case errors.Is(err, service.ErrNoFieldsToUpdate):
		return http.StatusBadRequest, "no valid fields provided to update"
	case errors.Is(err, service.ErrInvalidFileType):
		return http.StatusUnsupportedMediaType, "only jpeg, png, and webp images are accepted"
	case errors.Is(err, service.ErrFileTooLarge):
		return http.StatusRequestEntityTooLarge, "image must be smaller than 5 MB"
	case errors.Is(err, repository.ErrUserNotFound):
		return http.StatusNotFound, "user not found"
	default:
		return http.StatusInternalServerError, "internal server error"
	}
}