package repository

import (
	"context"
	"errors"
	"time"


	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/developer-badhan/Flixor/internal/model"
)

/**
 * StreamRepository defines the contract for all streaming-related DB operations.
 * Using an interface makes it easy to mock in tests.
*/
type StreamRepository interface {
	GetMovieAndIncrementView(ctx context.Context, movieID primitive.ObjectID) (*model.Movie, error)
}

// streamRepository is the concrete MongoDB implementation.
type streamRepository struct {
	collection *mongo.Collection
}

// NewStreamRepository creates a new StreamRepository backed by MongoDB.
func NewStreamRepository(db *mongo.Database) StreamRepository {
	return &streamRepository{
		collection: db.Collection("movies"),
	}
}

/**
 * GetMovieAndIncrementView does two things atomically in a single MongoDB call:
 * 1. Finds the movie by ID
 * 2. Increments its view_count by 1
 * 
 * We use FindOneAndUpdate with ReturnDocument=After so we get the
 * updated document (with the new view count) in a single round-trip.
 * This is safe under high concurrency — MongoDB's $inc is atomic.
*/
func (r *streamRepository) GetMovieAndIncrementView(ctx context.Context, movieID primitive.ObjectID) (*model.Movie, error) {
	// Build filter: match by _id
	filter := bson.M{"_id": movieID}

	// Build update: atomically increment view_count by 1, update updated_at timestamp
	update := bson.M{
		"$inc": bson.M{"view_count": 1},
		"$set": bson.M{"updated_at": time.Now()},
	}

	/**
	 * ReturnDocument: After → return the document AFTER the update is applied
	 * This gives us the latest view_count in the response.
	*/
	opts := options.FindOneAndUpdate().
		SetReturnDocument(options.After)

	var movie model.Movie
	err := r.collection.FindOneAndUpdate(ctx, filter, update, opts).Decode(&movie)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			// Return a clean sentinel so the service layer can give a 404
			return nil, ErrMovieNotFound
		}
		return nil, err
	}

	return &movie, nil
}

/**
 * ErrMovieNotFound is a sentinel error for "movie does not exist".
 * Defining it in the repository package keeps error handling consistent
 * across the entire service layer.
*/
var ErrMovieNotFound = errors.New("movie not found")
