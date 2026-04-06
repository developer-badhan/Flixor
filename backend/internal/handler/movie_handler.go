package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/developer-badhan/Flixor/internal/service"

	"github.com/gin-gonic/gin"
)

/**
 * MovieHandler holds a reference to the movie service.
 * This is the only dependency this layer has — no DB, no IA API.
*/
type MovieHandler struct {
	svc service.MovieService
}

// NewMovieHandler creates a MovieHandler wired to the given MovieService.
func NewMovieHandler(svc service.MovieService) *MovieHandler {
	return &MovieHandler{svc: svc}
}

/**
 * GetMovies handles GET /movies with pagination.
 *
 * Query params:
 *   - page (default 1)
 *   - limit (default 20)
 *
 * Response shape:
 *	{
 *	  "data": [ ...movies ],
 *	  "total": 123,
 *	  "page": 1,
 *	  "limit": 20,
 *	  "total_pages": 7
 *	}
*/
func (h *MovieHandler) GetMovies(c *gin.Context) {
	page  := queryInt(c, "page",  1)
	limit := queryInt(c, "limit", 20)

	movies, total, err := h.svc.GetMovies(c.Request.Context(), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	totalPages := (total + int64(limit) - 1) / int64(limit) // ceiling division

	c.JSON(http.StatusOK, gin.H{
		"data":        movies,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": totalPages,
	})
}

/** 
 * GetMovieByID handles GET /movies/:id to fetch a single movie by its ID.
 *
 * Path param:
 *   - id (string, required)
 *
 * Response shape:
 *	{
 *	  "data": { ...movie }
 *	}
 * Error handling:
 *   - 400 Bad Request if ID is missing or invalid format
 *   - 404 Not Found if movie is not found
*/
func (h *MovieHandler) GetMovieByID(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "movie id is required"})
		return
	}

	movie, err := h.svc.GetMovieByID(c.Request.Context(), id)
	if err != nil {
		// Differentiate error types for proper HTTP status codes
		switch {
		case errors.Is(err, service.ErrInvalidID):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid movie id format"})
		case errors.Is(err, service.ErrMovieNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "movie not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch movie"})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": movie})
}

/** 
 * SyncMovies handles POST /movies/sync to trigger a sync with the external movie API.
 *
 * Query params:
 *   - rows (default 50): number of movies to sync per page
 *   - page (default 1): which page of results to sync
 *
 * Response shape:
 *	{
 *	  "message": "sync complete",
 *	  "synced": 123
 *	}
 * Error handling:
 *   - 400 Bad Request if query params are invalid
 *   - 500 Internal Server Error if sync fails
*/
func (h *MovieHandler) SyncMovies(c *gin.Context) {
	rows := queryInt(c, "rows", 50)
	page := queryInt(c, "page", 1)

	count, err := h.svc.SyncMovies(c.Request.Context(), rows, page)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "sync failed: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "sync complete",
		"synced":  count,
	})
}

/**
 * queryInt reads an integer query param with a fallback default.
 * If the param is missing, empty, non-integer, or <= 0, it returns the default value.
*/
func queryInt(c *gin.Context, key string, defaultVal int) int {
	raw := c.Query(key)
	if raw == "" {
		return defaultVal
	}
	val, err := strconv.Atoi(raw)
	if err != nil || val <= 0 {
		return defaultVal
	}
	return val
}
