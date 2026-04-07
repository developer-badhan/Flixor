package handler

import (
	"errors"
	"net/http"

	"github.com/developer-badhan/Flixor/internal/service"

	"github.com/gin-gonic/gin"
)

/**
 * StreamHandler holds the dependencies needed to handle streaming HTTP requests.
 * It has a single method GetStream which handles the GET /stream/:id endpoint.
 * All business logic is delegated to the StreamService, keeping the handler focused on HTTP concerns.
*/
type StreamHandler struct {
	streamService service.StreamService
}

/** 
 * NewStreamHandler is a constructor function that initializes a StreamHandler with the given StreamService.
 * This promotes loose coupling and makes it easier to test the handler by injecting mock services.
 * @param streamService The service responsible for stream-related business logic
 * @return A pointer to a new StreamHandler instance
 * Example usage:
	streamService := service.NewStreamService(movieRepo, streamRepo)
	streamHandler := handler.NewStreamHandler(streamService)
	
*/
func NewStreamHandler(streamService service.StreamService) *StreamHandler {
	return &StreamHandler{
		streamService: streamService,
	}
}

/**
 * GetStream handles: GET /stream/:id
 *
 * This endpoint:
 *   - Requires JWT authentication (enforced by middleware in the router)
 *   - Reads the movie ID from the URL path
 *   - Delegates all logic to StreamService
 *   - Returns the stream URL + metadata + updated view count
 *
 * Response shape:
 * 
 * 200 OK → { success: true, data: { movie_id, title, stream_url, ... } }
 * 400 Bad Request → invalid movie ID format
 * 404 Not Found → movie doesn't exist
 * 422 Unprocessable → movie exists but has no video URL
 * 500 Internal Server Error → unexpected DB/server error
 */
func (h *StreamHandler) GetStream(c *gin.Context) {
	// Extract the movie ID from the URL path parameter
	movieID := c.Param("id")
	if movieID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "movie id is required",
		})
		return
	}

	// Call the service — all business logic lives there
	streamInfo, err := h.streamService.GetStreamInfo(c.Request.Context(), movieID)
	if err != nil {
		h.handleStreamError(c, err)
		return
	}

	// Success: return the stream info
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    streamInfo,
	})
}

/**
 * handleStreamError maps service-layer errors to appropriate HTTP responses
 * Centralizing this keeps GetStream clean and each error type handled consistently.
 * @param c The Gin context to write the response to
 * @param err The error returned by the service layer
 * Example usage:
 *	err := service.ErrMovieNotFound
 *	h.handleStreamError(c, err)
 *	handleStreamError will then send a 404 response with a JSON body indicating the movie was not found.
 */
func (h *StreamHandler) handleStreamError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidMovieID):
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "invalid movie id format",
		})

	case errors.Is(err, service.ErrMovieNotFound):
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "movie not found",
		})

	case errors.Is(err, service.ErrStreamNotAvailable):
		// Movie exists, but no video URL — this is a data issue, not a client error
		c.JSON(http.StatusUnprocessableEntity, gin.H{
			"success": false,
			"error":   "stream not available for this movie",
		})

	default:
		// Unexpected error — log it server-side, return generic message to client
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "internal server error",
		})
	}
}
