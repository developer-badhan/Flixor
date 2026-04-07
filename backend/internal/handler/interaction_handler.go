package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/developer-badhan/Flixor/internal/model"
	"github.com/developer-badhan/Flixor/internal/service"
)

/**
 * InteractionHandler wires HTTP routes to the interaction service.
 * It handles the HTTP layer of the interaction service.
 * Get user ID from JWT context.
 * Handle interaction errors.
 * Handle watchlist operations.
 * Handle reaction operations.
*/
type InteractionHandler struct {
	svc service.InteractionService
}

// NewInteractionHandler creates the handler with its service dependency.
func NewInteractionHandler(svc service.InteractionService) *InteractionHandler {
	return &InteractionHandler{svc: svc}
}

/**
 * getUserID pulls the authenticated user's ID string that the auth middleware
 * injects into the Gin context after validating the JWT.
*/
func getUserID(c *gin.Context) (string, bool) {
	userID, exists := c.Get("userID") 
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return "", false
	}
	return userID.(string), true
}

/**
 * handleInteractionError maps service-layer sentinel errors to HTTP status codes.
 * Centralizing this avoids repetitive if/else chains in every handler.
*/
func handleInteractionError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidMovieID):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrInvalidReaction):
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
	case errors.Is(err, service.ErrNoReactionExists):
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
	default:
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}

/**
 *  AddToWatchlist godoc
 *  @Summary      Add a movie to the watchlist
 *  @Tags         watchlist
 *  @Security     BearerAuth
 *  @Param        movieId  path  string  true  "Movie ID"
 *  @Success      200  {object}  map[string]string
 *  @Router       /watchlist/{movieId} [post]
*/
func (h *InteractionHandler) AddToWatchlist(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	movieID := c.Param("movieId")
	if movieID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "movieId is required"})
		return
	}

	if err := h.svc.AddToWatchlist(c.Request.Context(), userID, movieID); err != nil {
		handleInteractionError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "movie added to watchlist"})
}

/**
 * RemoveFromWatchlist godoc
 * @Summary      Remove a movie from the watchlist
 * @Tags         watchlist
 * @Security     BearerAuth
 * @Param        movieId  path  string  true  "Movie ID"
 * @Success      200  {object}  map[string]string
 * @Router       /watchlist/{movieId} [delete]
*/
func (h *InteractionHandler) RemoveFromWatchlist(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	movieID := c.Param("movieId")
	if movieID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "movieId is required"})
		return
	}

	if err := h.svc.RemoveFromWatchlist(c.Request.Context(), userID, movieID); err != nil {
		handleInteractionError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "movie removed from watchlist"})
}

/**
 * GetWatchlist godoc
 * @Summary      Get the authenticated user's watchlist
 * @Tags         watchlist
 * @Security     BearerAuth
 * @Success      200  {object}  model.Watchlist
 * @Router       /watchlist [get]
*/
func (h *InteractionHandler) GetWatchlist(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	watchlist, err := h.svc.GetWatchlist(c.Request.Context(), userID)
	if err != nil {
		handleInteractionError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"watchlist": watchlist.Movies,
		"count":     len(watchlist.Movies),
	})
}

/**
 * ReactToMovie godoc
 * @Summary      Like or dislike a movie
 * @Tags         reactions
 * @Security     BearerAuth
 * @Param        movieId  path    string           true  "Movie ID"
 * @Param        body     body    model.ReactRequest  true  "Reaction body"
 * @Success      200  {object}  map[string]string
 * @Router       /movies/{movieId}/react [post]
*/
func (h *InteractionHandler) ReactToMovie(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	movieID := c.Param("movieId")
	if movieID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "movieId is required"})
		return
	}

	var req model.ReactRequest
	// ShouldBindJSON returns 400 automatically if binding fails due to the
	// `binding:"required,oneof=like dislike"` tags on ReactRequest.
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "reaction must be 'like' or 'dislike'"})
		return
	}

	if err := h.svc.ReactToMovie(c.Request.Context(), userID, movieID, req.Reaction); err != nil {
		handleInteractionError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "reaction saved",
		"reaction": req.Reaction,
	})
}

/**
 * RemoveReaction godoc
 * @Summary      Remove reaction from a movie
 * @Tags         reactions
 * @Security     BearerAuth
 * @Param        movieId  path  string  true  "Movie ID"
 * @Success      200  {object}  map[string]string
 * @Router       /movies/{movieId}/react [delete]
*/
func (h *InteractionHandler) RemoveReaction(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	movieID := c.Param("movieId")
	if movieID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "movieId is required"})
		return
	}

	if err := h.svc.RemoveReaction(c.Request.Context(), userID, movieID); err != nil {
		handleInteractionError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "reaction removed"})
}

/**
 * GetReaction godoc
 * @Summary      Get your reaction on a specific movie
 * @Tags         reactions
 * @Security     BearerAuth
 * @Param        movieId  path  string  true  "Movie ID"
 * @Success      200  {object}  map[string]interface{}
 * @Router       /movies/{movieId}/reaction [get]
*/
func (h *InteractionHandler) GetReaction(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	movieID := c.Param("movieId")
	if movieID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "movieId is required"})
		return
	}

	reaction, err := h.svc.GetReaction(c.Request.Context(), userID, movieID)
	if err != nil {
		handleInteractionError(c, err)
		return
	}

	// Nil reaction means the user hasn't reacted yet — return neutral state.
	if reaction == nil {
		c.JSON(http.StatusOK, gin.H{
			"reacted":  false,
			"reaction": nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"reacted":    true,
		"reaction":   reaction.Type,
		"updated_at": reaction.UpdatedAt,
	})
}


/**
 * RecordWatch godoc
 * @Summary      Record that the user watched a movie
 * @Tags         history
 * @Security     BearerAuth
 * @Param        movieId  path  string  true  "Movie ID"
 * @Success      200  {object}  map[string]string
 * @Router       /history/{movieId} [post]
*/
func (h *InteractionHandler) RecordWatch(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	movieID := c.Param("movieId")
	if movieID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "movieId is required"})
		return
	}

	if err := h.svc.RecordWatch(c.Request.Context(), userID, movieID); err != nil {
		handleInteractionError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "watch recorded"})
}

/**
 * GetHistory godoc
 * @Summary      Get the authenticated user's watch history
 * @Tags         history
 * @Security     BearerAuth
 * @Success      200  {object}  map[string]interface{}
 * @Router       /history [get]
*/
func (h *InteractionHandler) GetHistory(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	history, err := h.svc.GetHistory(c.Request.Context(), userID)
	if err != nil {
		handleInteractionError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"history": history.Events,
		"count":   len(history.Events),
	})
}

/**
 * ClearHistory godoc
 * @Summary      Clear the authenticated user's watch history
 * @Tags         history
 * @Security     BearerAuth
 * @Success      200  {object}  map[string]string
 * @Router       /history [delete]
*/
func (h *InteractionHandler) ClearHistory(c *gin.Context) {
	userID, ok := getUserID(c)
	if !ok {
		return
	}

	if err := h.svc.ClearHistory(c.Request.Context(), userID); err != nil {
		handleInteractionError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "watch history cleared"})
}