package repository

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/developer-badhan/Flixor/internal/model"
)

// Interface — the service depends on this, never on the concrete struct.

/**
 * MovieRepository defines the methods for interacting with movie data in MongoDB.
 * The service layer depends on this interface, not the concrete implementation.
 * This allows for easier testing and future flexibility (e.g. swapping databases).
 * Methods include:
 * - BulkUpsert: Insert or update movies by their IA identifier (idempotent).
 * - FindAll: Get a paginated list of movies.
 * - FindByID: Get a movie by its MongoDB ObjectID.
 * - FindByIdentifier: Get a movie by its Internet Archive identifier.
 * - CountAll: Get the total number of movies (for pagination metadata).
 */
type MovieRepository interface {
	// BulkUpsert inserts or updates movies by their IA identifier.
	BulkUpsert(ctx context.Context, movies []model.Movie) error

	// FindAll returns a paginated list of movies.
	FindAll(ctx context.Context, skip, limit int64) ([]model.Movie, error)

	// FindByID returns one movie by its MongoDB ObjectID.
	FindByID(ctx context.Context, id primitive.ObjectID) (*model.Movie, error)

	// FindByIdentifier returns one movie by its Internet Archive identifier.
	FindByIdentifier(ctx context.Context, identifier string) (*model.Movie, error)

	// CountAll returns the total number of movies (for pagination metadata).
	CountAll(ctx context.Context) (int64, error)

	// FindByGenre returns a paginated list of movies filtered by genre (case-insensitive),
	// with the total count of matching movies for pagination metadata.
	FindByGenre(ctx context.Context, genre string, skip, limit int64) ([]model.Movie, int64, error)

	// SearchMovies returns a paginated list of movies based on the provided filter.
	SearchMovies(ctx context.Context, filter model.SearchFilter) (model.PaginatedMovies, error)
}

// Concrete implementation

/**
 * movieRepository is the concrete implementation of MovieRepository using MongoDB.
 * It holds a reference to the MongoDB collection and implements all methods defined in the interface.
 * The constructor ensures a unique index on the "identifier" field to prevent duplicates.
 * Each method interacts with MongoDB using the official Go driver, handling errors appropriately.
 * The service layer should only depend on the MovieRepository interface, not this struct.
 */
type movieRepository struct {
	col *mongo.Collection
}

// NewMovieRepository creates a new repository wired to the given collection.
func NewMovieRepository(db *mongo.Database) MovieRepository {
	col := db.Collection("movies")

	// Create a unique index on "identifier" so we never store duplicates.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	textIndex := mongo.IndexModel{
		Keys: bson.D{
			{Key: "title", Value: "text"},
			{Key: "description", Value: "text"},
		},
	}
	_, _ = col.Indexes().CreateOne(ctx, textIndex)

	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "identifier", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, _ = col.Indexes().CreateOne(ctx, indexModel) // non-fatal if already exists

	return &movieRepository{col: col}
}

// Method implementations

/**
 * BulkUpsert uses MongoDB's UpdateOne with upsert=true for each movie.
 * "Upsert" means: update if exists, insert if not. This is idempotent.
 */
func (r *movieRepository) BulkUpsert(ctx context.Context, movies []model.Movie) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()	
	if len(movies) == 0 {
		return nil
	}

	for _, movie := range movies {
		filter := bson.M{"identifier": movie.Identifier}

		/**
		 * $setOnInsert only sets CreatedAt the very first time (on insert).
		 * $set updates mutable fields every time.
		 */
		update := bson.M{
			"$set": bson.M{
				"title":         movie.Title,
				"description":   movie.Description,
				"year":          movie.Year,
				"genres":        movie.Genres,
				"director":      movie.Director,
				"thumbnail_url": movie.ThumbnailURL,
				"stream_url":    movie.StreamURL,
				"updated_at":    time.Now(),
			},
			"$setOnInsert": bson.M{
				"identifier": movie.Identifier,
				"view_count": 0,
				"created_at": time.Now(),
			},
		}

		opts := options.Update().SetUpsert(true)
		_, err := r.col.UpdateOne(ctx, filter, update, opts)
		if err != nil {
			return fmt.Errorf("movie repo: upsert failed for %q: %w", movie.Identifier, err)
		}
	}

	return nil
}


/**
 * FindAll returns movies with skip/limit for pagination.
 * Results are sorted by title alphabetically for a consistent UX.
 */
func (r *movieRepository) FindAll(ctx context.Context, skip, limit int64) ([]model.Movie, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.D{{Key: "title", Value: 1}})

	cursor, err := r.col.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, fmt.Errorf("movie repo: find all failed: %w", err)
	}
	defer cursor.Close(ctx)

	var movies []model.Movie
	if err := cursor.All(ctx, &movies); err != nil {
		return nil, fmt.Errorf("movie repo: cursor decode failed: %w", err)
	}

	return movies, nil
}

// FindByID fetches a single movie by its MongoDB _id.
func (r *movieRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*model.Movie, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()	
	var movie model.Movie
	err := r.col.FindOne(ctx, bson.M{"_id": id}).Decode(&movie)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil // not found → return nil, nil (service handles this)
		}
		return nil, fmt.Errorf("movie repo: find by id failed: %w", err)
	}
	return &movie, nil
}

// FindByIdentifier fetches a movie by its Internet Archive identifier string.
func (r *movieRepository) FindByIdentifier(ctx context.Context, identifier string) (*model.Movie, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()	
	var movie model.Movie
	err := r.col.FindOne(ctx, bson.M{"identifier": identifier}).Decode(&movie)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, fmt.Errorf("movie repo: find by identifier failed: %w", err)
	}
	return &movie, nil
}

// CountAll returns the total document count for pagination metadata.
func (r *movieRepository) CountAll(ctx context.Context) (int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()	
	count, err := r.col.CountDocuments(ctx, bson.M{})
	if err != nil {
		return 0, fmt.Errorf("movie repo: count failed: %w", err)
	}
	return count, nil
}

// FindByGenre returns movies filtered by genre with pagination.
// Genre filtering is case-insensitive using regex.
func (r *movieRepository) FindByGenre(ctx context.Context, genre string, skip, limit int64) ([]model.Movie, int64, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()	
	// Build case-insensitive regex filter for genre
	query := bson.M{
		"genres": bson.M{
			"$elemMatch": bson.M{
				"$regex": genre,
				"$options": "i",
			},
		},
	}

	// Count total matching documents
	total, err := r.col.CountDocuments(ctx, query)
	if err != nil {
		return nil, 0, fmt.Errorf("movie repo: count by genre failed: %w", err)
	}

	// Find options (sort + skip + limit)
	findOpts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetSort(bson.D{{Key: "title", Value: 1}}) // sort by title alphabetically

	// Fetch paginated results
	cursor, err := r.col.Find(ctx, query, findOpts)
	if err != nil {
		return nil, 0, fmt.Errorf("movie repo: find by genre failed: %w", err)
	}
	defer cursor.Close(ctx)

	var movies []model.Movie
	if err := cursor.All(ctx, &movies); err != nil {
		return nil, 0, fmt.Errorf("movie repo: cursor decode failed: %w", err)
	}

	// Guard: never return nil slice — frontend prefers an empty array
	if movies == nil {
		movies = []model.Movie{}
	}

	return movies, total, nil
}

// SearchMovies executes a dynamic query built from SearchFilter.
func (r *movieRepository) SearchMovies(ctx context.Context, filter model.SearchFilter) (model.PaginatedMovies, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()	
	// 1. Build the BSON filter
	query := bson.D{}

	if filter.Title != "" {
		query = append(query, bson.E{
			Key:   "$text",
			Value: bson.D{{Key: "$search", Value: filter.Title}},
		})
	}

	if filter.Genre != "" {
		// Case-insensitive regex so "action", "Action", "ACTION" all match.
		query = append(query, bson.E{
			Key: "genres",
			Value: bson.D{
				{Key: "$regex", Value: filter.Genre},
				{Key: "$options", Value: "i"},
			},
		})
	}

	// 2. Pagination maths
	skip := int64((filter.Page - 1) * filter.Limit)
	if skip > 10000 {
		return model.PaginatedMovies{}, fmt.Errorf("page too large")
	}

	// 3. Find options (sort + skip + limit)
	findOpts := options.Find().
		SetSkip(skip).
		SetLimit(int64(filter.Limit))

	// When the user searched by title, sort by text-relevance score (best match first).
	// Otherwise fall back to newest-first (descending _id is a cheap proxy for insert time).
	if filter.Title != "" {
		findOpts.SetProjection(bson.D{
			{Key: "score", Value: bson.D{{Key: "$meta", Value: "textScore"}}},
		})
		findOpts.SetSort(bson.D{
			{Key: "score", Value: bson.D{{Key: "$meta", Value: "textScore"}}},
			{Key: "_id", Value: -1},
		})
	} else {
		findOpts.SetSort(bson.D{{Key: "_id", Value: -1}})
	}

	// 4. Count total matching documents
	total, err := r.col.CountDocuments(ctx, query)
	if err != nil {
		return model.PaginatedMovies{}, err
	}

	// 5. Fetch the page
	cursor, err := r.col.Find(ctx, query, findOpts)
	if err != nil {
		return model.PaginatedMovies{}, err
	}
	defer cursor.Close(ctx)

	var movies []model.Movie
	if err := cursor.All(ctx, &movies); err != nil {
		return model.PaginatedMovies{}, err
	}

	// Guard: never return nil slice — frontend prefers an empty array
	if movies == nil {
		movies = []model.Movie{}
	}

	// 6. Build response
	pages := int64(math.Ceil(float64(total) / float64(filter.Limit)))

	return model.PaginatedMovies{
		Total:  total,
		Page:   filter.Page,
		Limit:  filter.Limit,
		Pages:  pages,
		Movies: movies,
	}, nil
}
