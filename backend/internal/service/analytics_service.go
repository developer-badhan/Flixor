package service

import (
	"context"
	"fmt"
	"time"

	"github.com/developer-badhan/Flixor/internal/model"
	"github.com/developer-badhan/Flixor/internal/repository"
)

// Defaults & Constants
const (
	defaultAnalyticsLimit = 10
	maxAnalyticsLimit     = 100
)

/**
 * windowDurations maps the window string param to a Go duration.
 * We support three windows — short, medium, long.
*/
var windowDurations = map[string]time.Duration{
	"1d":  24 * time.Hour,
	"7d":  7 * 24 * time.Hour,
	"30d": 30 * 24 * time.Hour,
}

/**
 * windowLabels is the human-readable label for each window (used in API response).
*/
var windowLabels = map[string]string{
	"1d":  "Last 24 hours",
	"7d":  "Last 7 days",
	"30d": "Last 30 days",
}

/**
 * AnalyticsService: holds the analytics repository
 * It is the bridge between the controller and the repository
*/
type AnalyticsService struct {
	repo *repository.AnalyticsRepository
}

/**
 * NewAnalyticsService: creates a new instance of AnalyticsService
*/
func NewAnalyticsService(repo *repository.AnalyticsRepository) *AnalyticsService {
	return &AnalyticsService{repo: repo}
}

/**
 * GetTrending: returns movies ranked by views in the requested time window.
 * Handles defaults, parses the window string, then delegates to the repository.
*/
func (s *AnalyticsService) GetTrending(ctx context.Context, q model.TrendingQuery) (*model.TrendingResponse, error) {
	// Apply defaults
	window := q.Window
	if window == "" {
		window = "7d"
	}
	limit := clampLimit(q.Limit)

	// Resolve the window duration
	duration, ok := windowDurations[window]
	if !ok {
		// Should never happen because the handler validates via `oneof`, but be safe
		return nil, fmt.Errorf("analytics: unknown window %q", window)
	}

	since := time.Now().UTC().Add(-duration)

	// Fetch from DB
	dbResults, err := s.repo.GetTrending(ctx, since, limit)
	if err != nil {
		return nil, fmt.Errorf("analytics: trending query: %w", err)
	}

	// Map DB results → API model, adding 1-based rank numbers
	movies := make([]model.TrendingMovie, 0, len(dbResults))
	for i, r := range dbResults {
		movies = append(movies, model.TrendingMovie{
			ID:            r.ID.Hex(),
			Title:         r.Title,
			Genre:         r.Genre,
			Thumbnail:     r.Thumbnail,
			Year:          r.Year,
			ViewsInWindow: r.ViewsInWindow,
			TotalViews:    r.TotalViews,
			Rank:          i + 1, // rank is 1-indexed, humans don't count from 0
		})
	}

	return &model.TrendingResponse{
		Window:       window,
		WindowLabel:  windowLabels[window],
		TotalResults: len(movies),
		Movies:       movies,
	}, nil
}

/**
 * GetMostWatched: returns movies sorted by all-time view_count.
 * This is different from GetTrending because it does not have a time window.
 * It is used to get the most watched movies of all time.
*/
func (s *AnalyticsService) GetMostWatched(ctx context.Context, q model.TopQuery) (*model.MostWatchedResponse, error) {
	limit := clampLimit(q.Limit)

	dbResults, err := s.repo.GetMostWatched(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("analytics: most-watched query: %w", err)
	}

	movies := make([]model.MostWatchedMovie, 0, len(dbResults))
	for i, r := range dbResults {
		movies = append(movies, model.MostWatchedMovie{
			ID:         r.ID.Hex(),
			Title:      r.Title,
			Genre:      r.Genre,
			Thumbnail:  r.Thumbnail,
			Year:       r.Year,
			TotalViews: r.TotalViews,
			Rank:       i + 1,
		})
	}

	return &model.MostWatchedResponse{
		TotalResults: len(movies),
		Movies:       movies,
	}, nil
}

/**
 * GetTopGenres: returns genres ranked by total views across all movies.
 * This is different from GetTrending and GetMostWatched because it returns genres instead of movies.
 * It is used to get the most watched genres of all time.
*/
func (s *AnalyticsService) GetTopGenres(ctx context.Context, q model.TopQuery) (*model.TopGenresResponse, error) {
	limit := clampLimit(q.Limit)

	dbResults, err := s.repo.GetTopGenres(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("analytics: top-genres query: %w", err)
	}

	// Map DB results → API model
	genres := make([]model.GenreStat, 0, len(dbResults))
	for i, r := range dbResults {
		genres = append(genres, model.GenreStat{
			Genre:      r.Genre,
			MovieCount: r.MovieCount,
			TotalViews: r.TotalViews,
			Rank:       i + 1,
		})
	}

	return &model.TopGenresResponse{
		TotalResults: len(genres),
		Genres:       genres,
	}, nil
}

/**
 * GetPlatformStats: returns a snapshot of platform-wide numbers.
 * This is different from GetTrending, GetMostWatched, and GetTopGenres because it returns platform-wide numbers instead of movie or genre numbers.
 * It is used to get the platform-wide numbers.
*/
func (s *AnalyticsService) GetPlatformStats(ctx context.Context) (*model.PlatformStats, error) {
	raw, err := s.repo.GetPlatformStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("analytics: platform stats query: %w", err)
	}

	return &model.PlatformStats{
		TotalMovies:    raw.TotalMovies,
		TotalUsers:     raw.TotalUsers,
		TotalViews:     raw.TotalViews,
		TotalWatchlist: raw.TotalWatchlist,
		TotalLikes:     raw.TotalLikes,
	}, nil
}

/**
 * clampLimit: applies default and maximum bounds to the limit parameter.
 * This is a pure function — no side effects, easy to test.
*/
func clampLimit(limit int) int {
	if limit <= 0 {
		return defaultAnalyticsLimit
	}
	if limit > maxAnalyticsLimit {
		return maxAnalyticsLimit
	}
	return limit
}