package utils

import "github.com/gin-gonic/gin"

/**
 * APIResponse is the standard envelope for every API response.
 * Keeping a consistent shape makes frontend integration predictable.
 *
 *	{
 *	  "success": true,
 *	  "message": "Movies fetched successfully",
 *	  "data":    { ... }
 *	}
*/
type APIResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    any `json:"data,omitempty"`
}

/**
 * APIErrorResponse is the standard envelope for error API responses.
 * Keeping a consistent shape makes frontend integration predictable.
 *
 *	{
 *	  "success": false,
 *	  "message": "Movie not found"
 *	}
*/
type APIErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

/**
 * SuccessResponse writes a 2xx JSON response.
 * The 'data' field is optional and will be omitted if nil.
 * Example usage:
 *
 *	SuccessResponse(c, http.StatusOK, "Movies fetched successfully", movies)
*/
func SuccessResponse(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, APIResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

/**
 * ErrorResponse writes a 4xx/5xx JSON response.
 * The 'data' field is omitted since it's an error response.
 * Example usage:
 *
 *	ErrorResponse(c, http.StatusNotFound, "Movie not found")
*/
func ErrorResponse(c *gin.Context, statusCode int, message string) {
	c.AbortWithStatusJSON(statusCode, APIErrorResponse{
		Success: false,
		Message: message,
	})
}
