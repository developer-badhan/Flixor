package repository

import (
	"context"
	"errors"
	"fmt"
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
	count, err := r.col.CountDocuments(ctx, bson.M{})
	if err != nil {
		return 0, fmt.Errorf("movie repo: count failed: %w", err)
	}
	return count, nil
}
