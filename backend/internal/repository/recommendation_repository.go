package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

/**
 * RecommendationRepository handles all DB reads needed by the recommendation engines.
 * It reads from three collections: watch_history, likes, and movies.
*/
type RecommendationRepository struct {
	historyCol *mongo.Collection
	likesCol   *mongo.Collection
	moviesCol  *mongo.Collection
}

/**
 * NewRecommendationRepository creates a new RecommendationRepository.
 * It initializes the RecommendationRepository with the given database.
 * This is used by the recommendation service to get data from the database.
*/	
func NewRecommendationRepository(db *mongo.Database) *RecommendationRepository {
	return &RecommendationRepository{
		historyCol: db.Collection("watch_history"),
		likesCol:   db.Collection("likes"),
		moviesCol:  db.Collection("movies"),
	}
}

/**
 * GetWatchedMovieIDs returns the list of movie IDs the user has already watched.
 * We use this to exclude already-seen movies from recommendations.
*/
func (r *RecommendationRepository) GetWatchedMovieIDs(ctx context.Context, userID primitive.ObjectID) ([]primitive.ObjectID, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"user_id": userID}
	cursor, err := r.historyCol.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	// We only need the movie_id field — projection keeps it fast
	type historyDoc struct {
		MovieID primitive.ObjectID `bson:"movie_id"`
	}

	var ids []primitive.ObjectID
	for cursor.Next(ctx) {
		var doc historyDoc
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		ids = append(ids, doc.MovieID)
	}
	return ids, cursor.Err()
}

/**
 * GetLikedMovieIDs returns movie IDs the user has explicitly liked.
 * Liked movies carry stronger genre signals than watch history.
*/
func (r *RecommendationRepository) GetLikedMovieIDs(ctx context.Context, userID primitive.ObjectID) ([]primitive.ObjectID, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	filter := bson.M{"user_id": userID, "action": "like"}
	cursor, err := r.likesCol.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	type likeDoc struct {
		MovieID primitive.ObjectID `bson:"movie_id"`
	}

	var ids []primitive.ObjectID
	for cursor.Next(ctx) {
		var doc likeDoc
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		ids = append(ids, doc.MovieID)
	}
	return ids, cursor.Err()
}

/**
 * GetGenresFromMovieIDs fetches the genres for a set of movie IDs.
 * Returns a frequency map: genre → count. Higher count = stronger signal.
*/
func (r *RecommendationRepository) GetGenresFromMovieIDs(ctx context.Context, movieIDs []primitive.ObjectID) (map[string]int, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if len(movieIDs) == 0 {
		return map[string]int{}, nil
	}

	filter := bson.M{"_id": bson.M{"$in": movieIDs}}
	opts := options.Find().SetProjection(bson.M{"genre": 1})

	cursor, err := r.moviesCol.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	type movieDoc struct {
		Genre []string `bson:"genre"`
	}

	// genreFreq is a map that stores the frequency of each genre.
	// Higher frequency means the user has watched more movies of that genre.
	genreFreq := make(map[string]int)
	for cursor.Next(ctx) {
		var doc movieDoc
		if err := cursor.Decode(&doc); err != nil {
			continue
		}
		for _, g := range doc.Genre {
			genreFreq[g]++
		}
	}
	return genreFreq, cursor.Err()
}

/**
 * GetMoviesByGenres finds movies that match any of the given genres,
 * excludes already-watched movies, and sorts by popularity (view_count desc).
 * limit controls max results returned from DB.
*/
type RecoMovie struct {
	ID        primitive.ObjectID `bson:"_id"`
	Title     string             `bson:"title"`
	Genre     []string           `bson:"genre"`
	Thumbnail string             `bson:"thumbnail"`
	Year      string             `bson:"year"`
	ViewCount int64              `bson:"view_count"`
}

// GetMoviesByGenres finds movies that match any of the given genres,
// excludes already-watched movies, and sorts by popularity (view_count desc).
// limit controls max results returned from DB.
func (r *RecommendationRepository) GetMoviesByGenres(
	ctx context.Context,
	genres []string,
	excludeIDs []primitive.ObjectID,
	limit int,
) ([]RecoMovie, error) {
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	filter := bson.M{
		"genre": bson.M{"$in": genres},
	}
	if len(excludeIDs) > 0 {
		filter["_id"] = bson.M{"$nin": excludeIDs}
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "view_count", Value: -1}}).
		SetLimit(int64(limit))

	cursor, err := r.moviesCol.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var movies []RecoMovie
	if err := cursor.All(ctx, &movies); err != nil {
		return nil, err
	}
	return movies, nil
}

/**
 * GetMoviesByTitles fetches movies whose titles appear in the given list.
 * Used when Gemini returns a list of recommended movie titles — we match them in our DB.
 * Uses a case-insensitive regex match for robustness.
*/
func (r *RecommendationRepository) GetMoviesByTitles(
	ctx context.Context,
	titles []string,
	excludeIDs []primitive.ObjectID,
	limit int,
) ([]RecoMovie, error) {
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	if len(titles) == 0 {
		return nil, nil
	}

	// Build $or of regex conditions — each title matched case-insensitively
	orConditions := make([]bson.M, 0, len(titles))
	for _, t := range titles {
		orConditions = append(orConditions, bson.M{
			"title": bson.M{"$regex": t, "$options": "i"},
		})
	}

	filter := bson.M{"$or": orConditions}
	if len(excludeIDs) > 0 {
		filter["_id"] = bson.M{"$nin": excludeIDs}
	}

	opts := options.Find().
		SetLimit(int64(limit)).
		SetSort(bson.D{{Key: "view_count", Value: -1}})

	cursor, err := r.moviesCol.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var movies []RecoMovie
	if err := cursor.All(ctx, &movies); err != nil {
		return nil, err
	}
	return movies, nil
}