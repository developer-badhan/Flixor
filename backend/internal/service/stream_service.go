package service

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/developer-badhan/Flixor/internal/model"
	"github.com/developer-badhan/Flixor/internal/repository"
)

// StreamService defines the contract for all streaming business operations.
type StreamService interface {
	GetStreamInfo(ctx context.Context, movieID string) (*StreamResponse, error)
}

// StreamResponse is the data we return to the client when they request a stream.
// We don't expose the raw Movie model — we shape exactly what the frontend needs.
type StreamResponse struct {
	MovieID      string   `json:"movie_id"`
	Title        string   `json:"title"`
	StreamURL    string   `json:"stream_url"`
	ThumbnailURL string   `json:"thumbnail_url"`
	Genre        []string `json:"genre"`
	Year         string   `json:"year"`
	Director     string   `json:"director"`
	ViewCount    int64    `json:"view_count"` 
}

// streamService is the concrete implementation.
type streamService struct {
	streamRepo repository.StreamRepository
}

// NewStreamService creates a StreamService with its required dependencies injected.
func NewStreamService(streamRepo repository.StreamRepository) StreamService {
	return &streamService{
		streamRepo: streamRepo,
	}
}

// GetStreamInfo is the core business operation for Phase 3:
//  1. Validate the provided movie ID is a valid MongoDB ObjectID
//  2. Fetch the movie from DB and atomically increment its view count
//  3. Validate the movie actually has a streaming URL (some entries may be metadata-only)
//  4. Return a clean StreamResponse DTO
func (s *streamService) GetStreamInfo(ctx context.Context, movieID string) (*StreamResponse, error) {
	// Step 1: Parse & validate the ObjectID
	// This catches malformed IDs before any DB round-trip (cheap validation)
	objID, err := primitive.ObjectIDFromHex(movieID)
	if err != nil {
		return nil, ErrInvalidMovieID
	}

	// Step 2: Fetch movie + increment view count (single atomic DB operation)
	movie, err := s.streamRepo.GetMovieAndIncrementView(ctx, objID)
	if err != nil {
		if errors.Is(err, repository.ErrMovieNotFound) {
			// fallback: check if movie exists but has no stream
			existingMovie, findErr := s.streamRepo.FindByID(ctx, objID)

			if findErr == nil && existingMovie != nil {
				return nil, ErrStreamNotAvailable
			}

			if existingMovie == nil {
				return nil, ErrMovieNotFound
			}

			// Movie exists but has no stream URL
			return nil, ErrStreamNotAvailable
		}
		// Unexpected DB error — return as-is so the handler can log it
		return nil, err
	}

	// Step 3: Guard against movies that have no streaming URL.
	// Internet Archive occasionally has entries that are text-only or have broken links.
	if movie.StreamURL == "" {
		return nil, ErrStreamNotAvailable
	}

	// Step 4: Map the domain model → response DTO
	return mapMovieToStreamResponse(movie), nil
}

// mapMovieToStreamResponse converts a Movie model to a StreamResponse.
// Keeping this as a separate function makes it easy to unit test in isolation.
func mapMovieToStreamResponse(m *model.Movie) *StreamResponse {
	return &StreamResponse{
		MovieID:      m.ID.Hex(),
		Title:        m.Title,
		StreamURL:    m.StreamURL,
		ThumbnailURL: m.ThumbnailURL,
		Genre:        m.Genres,
		Year:         m.Year,
		Director:     m.Director,
		ViewCount:    m.ViewCount,
	}
}

// --- Service-level sentinel errors ---
// These are the errors the handler layer checks against to pick the right HTTP status.

// ErrInvalidMovieID is returned when the ID string is not a valid MongoDB ObjectID hex.
var ErrInvalidMovieID = errors.New("invalid movie id")

// ErrMovieNotFound is returned when no movie matches the given ID.
var ErrMovieNotFound = errors.New("movie not found")

// ErrStreamNotAvailable is returned when the movie exists but has no video URL.
var ErrStreamNotAvailable = errors.New("stream not available for this movie")
