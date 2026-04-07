package service

import (
	"context"
	"fmt"
	"time"

	"github.com/developer-badhan/Flixor/internal/model"
	"github.com/developer-badhan/Flixor/internal/repository"
	"github.com/developer-badhan/Flixor/pkg/utils"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Interface
/**
 * MovieService defines the business logic for movie-related operations.
 * It abstracts away the details of data fetching and storage, allowing
 * handlers to call high-level methods without worrying about implementation.
 * The service layer is where we enforce business rules and data transformations.
 */
type MovieService interface {

	// SyncMovies fetches movies from Internet Archive and upserts them into MongoDB.
	SyncMovies(ctx context.Context, rows, page int) (int, error)

	// GetMovies returns a paginated list of movies with total count.
	GetMovies(ctx context.Context, page, limit int) ([]model.Movie, int64, error)

	// GetMovieByID returns a single movie by its MongoDB ID string.
	GetMovieByID(ctx context.Context, id string) (*model.Movie, error)
}

// Concrete implementation
/** *
* movieService is the concrete implementation of MovieService.
* It has a dependency on MovieRepository, which it uses to interact with the database.
* The service layer should not know about HTTP or request/response details — it just provides methods that handlers can call.
 */
type movieService struct {
	repo repository.MovieRepository
}

// Constructor is called once at startup in main.go, with dependencies injected.
func NewMovieService(repo repository.MovieRepository) MovieService {
	return &movieService{repo: repo}
}

// Method implementations
/**
 * SyncMovies is the core ingestion pipeline:
 *  1. Fetch raw docs from Internet Archive
 *  2. Convert each IAMovie → model.Movie
 *  3. Bulk-upsert into MongoDB
 *  It returns the number of movies synced so the caller can log/report it.
 */
func (s *movieService) SyncMovies(ctx context.Context, rows, page int) (int, error) {
	// Clamp inputs
	if rows <= 0 || rows > 100 {
		rows = 50
	}
	if page <= 0 {
		page = 1
	}

	// Fetch from Internet Archive
	iaDocs, err := utils.FetchMoviesFromArchive(rows, page)
	if err != nil {
		return 0, fmt.Errorf("movie service: fetch from archive: %w", err)
	}

	if len(iaDocs) == 0 {
		return 0, nil
	}

	// Convert IAMovie → model.Movie
	movies := make([]model.Movie, 0, len(iaDocs))
	for _, doc := range iaDocs {
		movies = append(movies, model.Movie{
			Identifier:   doc.Identifier,
			Title:        doc.Title,
			Description:  doc.Description,
			Year:         doc.Year,
			Genres:       doc.Genres,
			Director:     doc.Director,
			ThumbnailURL: doc.ThumbnailURL,
			StreamURL:    doc.StreamURL,
			ViewCount:    0,
			CreatedAt:    time.Now(),
		})
	}

	// Upsert into MongoDB
	if err := s.repo.BulkUpsert(ctx, movies); err != nil {
		return 0, fmt.Errorf("movie service: bulk upsert: %w", err)
	}

	return len(movies), nil
}

/**
 * GetMovies returns a paginated movie list plus the total count.
 * The handler uses total count to build pagination metadata in the response.
 * Business rules enforced here:
 *  - Default page = 1, limit = 20 if invalid values provided
 *  - Max limit of 100 to prevent abuse
 */
func (s *movieService) GetMovies(ctx context.Context, page, limit int) ([]model.Movie, int64, error) {
	// Sanitize inputs
	if page <= 0 {
		page = 1
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	// Step 1: Check if DB is empty
	total, err := s.repo.CountAll(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("movie service: count movies: %w", err)
	}

	// Step 2: Lazy sync if empty
	if total == 0 {
		_, err := s.SyncMovies(ctx, 50, 1)
		if err != nil {
			return nil, 0, fmt.Errorf("movie service: initial sync failed: %w", err)
		}

		// Re-count after sync
		total, err = s.repo.CountAll(ctx)
		if err != nil {
			return nil, 0, fmt.Errorf("movie service: count after sync failed: %w", err)
		}
	}

	// Step 3: Fetch paginated data
	skip := int64((page - 1) * limit)

	movies, err := s.repo.FindAll(ctx, skip, int64(limit))
	if err != nil {
		return nil, 0, fmt.Errorf("movie service: get movies: %w", err)
	}

	return movies, total, nil
}

/**
 * GetMovieByID looks up a single movie by its MongoDB ObjectID string.
 * Returns a typed error if the ID format is invalid so the handler can
 * respond with 400 vs 404 vs 500 correctly.
 * GetMovieByID looks up a single movie by its MongoDB ObjectID string.
 * Returns a typed error if the ID format is invalid so the handler can
 * respond with 400 vs 404 vs 500 correctly.
 */
func (s *movieService) GetMovieByID(ctx context.Context, id string) (*model.Movie, error) {
	oid, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		// Return a special sentinel so the handler knows it's a bad request
		return nil, ErrInvalidID
	}

	movie, err := s.repo.FindByID(ctx, oid)
	if err != nil {
		return nil, fmt.Errorf("movie service: get by id: %w", err)
	}
	if movie == nil {
		return nil, ErrMovieNotFound
	}

	return movie, nil
}

// Sentinel errors — typed errors the handler can check with errors.Is()
var (
	ErrInvalidID = fmt.Errorf("invalid movie id format")
)
