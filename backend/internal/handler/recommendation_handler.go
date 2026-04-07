package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/developer-badhan/Flixor/internal/model"
	"github.com/developer-badhan/Flixor/internal/service"
)

/**
 * RecommendationHandler handles POST /api/v1/recommend.
 * It is protected by AuthMiddleware — userID is always available in context.
*/
type RecommendationHandler struct {
	recoService *service.RecommendationService
}

/**
 * NewRecommendationHandler creates a new RecommendationHandler.
 * It initializes the RecommendationHandler with the given recommendation service.
 * This is used by the recommendation service to get data from the database.
*/
func NewRecommendationHandler(recoService *service.RecommendationService) *RecommendationHandler {
	return &RecommendationHandler{recoService: recoService}
}

/**
 * Recommend godoc
 * @Summary      Get personalised movie recommendations
 * @Description  Returns movie recommendations using rule-based genre matching and/or Gemini AI.
 * @Tags         Recommendations
 * @Accept       json
 * @Produce      json
 * @Security     BearerAuth
 * @Param        body body model.RecommendRequest false "Recommendation options"
 * @Success      200  {object}  model.RecommendResponse
 * @Failure      400  {object}  map[string]string
 * @Failure      401  {object}  map[string]string
 * @Failure      500  {object}  map[string]string
 * @Router       /api/v1/recommend [post]
*/
func (h *RecommendationHandler) GetRecommendations(c *gin.Context) {
	/**
	 * 1. Parse request body (optional — all fields have defaults)
	 * This is where the recommendation service gets data from the database.
	 * ShouldBindJSON returns error only if body exists but is malformed.
	 * An empty body is fine; defaults will be used.
	*/
	var req model.RecommendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body: " + err.Error()})
		return
	}

	/**
	 * 2. Extract userID from JWT context (set by AuthMiddleware)
	 * We store the user ID as a hex string in the context key "userID".
	*/
	userIDHex, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized: missing user context"})
		return
	}

	userID, err := primitive.ObjectIDFromHex(userIDHex.(string))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid user ID in token"})
		return
	}

	/**
	 * 3. Delegate to service
	 * This is where the recommendation service gets data from the database.
	 * It uses the user ID and the request body to generate recommendations.
	*/
	resp, err := h.recoService.GetRecommendations(c.Request.Context(), userID, req)
	if err != nil {
		// Log the actual error server-side; surface a safe message to client
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  "could not generate recommendations",
			"detail": err.Error(), // include during development; remove in prod
		})
		return
	}

	/**
	 * 4. Edge case: no recommendations could be generated ──
	 * This happens when the user is brand new with no watch history
	 * and the AI engine also returned nothing matchable in our DB.
	*/
	if resp.TotalResults == 0 {
		c.JSON(http.StatusOK, gin.H{
			"message": "No recommendations yet. Watch a few movies to get personalised picks!",
			"movies":  []interface{}{},
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}