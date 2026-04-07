package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/developer-badhan/Flixor/internal/model"
	"github.com/developer-badhan/Flixor/internal/service"
)

/**
 * AnalyticsHandler handles all four analytics endpoints.
 * These endpoints are intentionally PUBLIC (no auth required) —
 * analytics data like trending and most-watched are typically surfaced
 * on the homepage for all visitors, logged in or not.
*/
type AnalyticsHandler struct {
	svc *service.AnalyticsService
}

/**
 * NewAnalyticsHandler: creates a new instance of AnalyticsHandler.
 * It takes a pointer to an AnalyticsService as an argument and returns a pointer to an AnalyticsHandler.
*/
func NewAnalyticsHandler(svc *service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{svc: svc}
}

/**
 * GetTrending godoc
 * @Summary      Get trending movies
 * @Description  Returns movies ranked by views in a recent time window.
 * @Tags         Analytics
 * @Produce      json
 * @Param        window  query  string  false  "Time window: 1d | 7d | 30d (default: 7d)"
 * @Param        limit   query  int     false  "Number of results (default: 10, max: 100)"
 * @Success      200  {object}  model.TrendingResponse
 * @Failure      400  {object}  map[string]string
 * @Failure      500  {object}  map[string]string
 * @Router       /api/v1/analytics/trending [get]
*/
func (h *AnalyticsHandler) GetTrending(c *gin.Context) {
	var q model.TrendingQuery

	// ShouldBindQuery binds URL query params (?window=7d&limit=10)
	// It validates "window" is one of 1d|7d|30d via `binding:"oneof=..."` tag
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":  "invalid query parameters",
			"detail": err.Error(),
		})
		return
	}

	resp, err := h.svc.GetTrending(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch trending movies"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

/**
 * GetMostWatched godoc
 * @Summary      Get all-time most watched movies
 * @Description  Returns movies ranked by their all-time total view count.
 * @Tags         Analytics
 * @Produce      json
 * @Param        limit  query  int  false  "Number of results (default: 10, max: 100)"
 * @Success      200  {object}  model.MostWatchedResponse
 * @Failure      400  {object}  map[string]string
 * @Failure      500  {object}  map[string]string
 * @Router       /api/v1/analytics/most-watched [get]
*/
func (h *AnalyticsHandler) GetMostWatched(c *gin.Context) {
	var q model.TopQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query parameters"})
		return
	}

	resp, err := h.svc.GetMostWatched(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch most-watched movies"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

/**
 * GetTopGenres godoc
 * @Summary      Get top genres by total views
 * @Description  Returns genres ranked by the sum of view counts across all their movies.
 * @Tags         Analytics
 * @Produce      json
 * @Param        limit  query  int  false  "Number of genres (default: 10, max: 100)"
 * @Success      200  {object}  model.TopGenresResponse
 * @Failure      400  {object}  map[string]string
 * @Failure      500  {object}  map[string]string
 * @Router       /api/v1/analytics/top-genres [get]
*/
func (h *AnalyticsHandler) GetTopGenres(c *gin.Context) {
	var q model.TopQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid query parameters"})
		return
	}

	resp, err := h.svc.GetTopGenres(c.Request.Context(), q)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch top genres"})
		return
	}

	c.JSON(http.StatusOK, resp)
}

/**
 * GetPlatformStats godoc
 * @Summary      Get platform-wide statistics
 * @Description  Returns aggregate counts: total movies, users, views, watchlists, likes.
 * @Tags         Analytics
 * @Produce      json
 * @Success      200  {object}  model.PlatformStats
 * @Failure      500  {object}  map[string]string
 * @Router       /api/v1/analytics/stats [get]
*/
func (h *AnalyticsHandler) GetPlatformStats(c *gin.Context) {
	stats, err := h.svc.GetPlatformStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "could not fetch platform stats"})
		return
	}

	c.JSON(http.StatusOK, stats)
}