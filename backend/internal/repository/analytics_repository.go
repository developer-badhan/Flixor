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
 * AnalyticsRepository runs all read-heavy analytics queries.
 * It touches five collections but NEVER writes — purely for reads.
*/
type AnalyticsRepository struct {
	moviesCol    *mongo.Collection
	historyCol   *mongo.Collection
	usersCol     *mongo.Collection
	watchlistCol *mongo.Collection
	reactionsCol *mongo.Collection // was: likesCol → "likes" (collection never existed)
}

/**
 * NewAnalyticsRepository creates a new AnalyticsRepository.
*/
func NewAnalyticsRepository(db *mongo.Database) *AnalyticsRepository {
	return &AnalyticsRepository{
		moviesCol:    db.Collection("movies"),
		historyCol:   db.Collection("watch_history"),
		usersCol:     db.Collection("users"),
		watchlistCol: db.Collection("watchlists"),
		reactionsCol: db.Collection("reactions"), // FIX 1: was db.Collection("likes")
	}
}

/**
 * TrendingMovieResult is what the aggregation pipeline produces.
 * We keep it internal to the repository package.
*/
type TrendingMovieResult struct {
	ID            primitive.ObjectID `bson:"_id"`
	Title         string             `bson:"title"`
	Genre         []string           `bson:"genre"`
	Thumbnail     string             `bson:"thumbnail"`
	Year          string             `bson:"year"`
	ViewsInWindow int64              `bson:"views_in_window"`
	TotalViews    int64              `bson:"total_views"`
}

/**
 * 1. TRENDING — Time-windowed aggregation
 * GetTrending returns movies ranked by views within a time window.
 *
 * Pipeline explained step by step:
 *
 *   Stage 1 — $unwind:
 *     Unwind the events array in the watch_history collection.
 *
 *   Stage 2 — $match:
 *     Filter watch_history entries to only those after `since` timestamp.
 *     This is the "trending window" — only recent watches count.
 *     MongoDB uses the index on `watched_at` for this → very fast.
 *
 *   Stage 3 — $group:
 *     Group remaining entries by movie_id.
 *     Count how many watch events each movie has in the window → views_in_window.
 *
 *   Stage 3 — $sort (intermediate):
 *     Sort by views_in_window descending so the hottest movie is first.
 *
 *   Stage 4 — $limit:
 *     Cut to the requested limit early, before the expensive $lookup.
 *     This is a critical performance optimisation — don't $lookup 10,000 movies.
 *
 *   Stage 5 — $lookup:
 *     Join each grouped result back to the movies collection to get
 *     title, genre, thumbnail, year, and the all-time view_count.
 *
 *   Stage 6 — $unwind:
 *     $lookup returns an array; $unwind flattens it to a single document.
 *     preserveNullAndEmptyArrays: false drops any movie_id not found in movies
 *     (safety net for orphaned history entries).
 *
 *   Stage 7 — $project:
 *     Shape the final output — pick only the fields we need.
*/
func (r *AnalyticsRepository) GetTrending(ctx context.Context, since time.Time, limit int) ([]TrendingMovieResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	pipeline := mongo.Pipeline{
		{{Key: "$unwind", Value: "$events"}},
		{{Key: "$match", Value: bson.M{
			"events.watched_at": bson.M{"$gte": since},
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":             "$events.movie_id",
			"views_in_window": bson.M{"$sum": 1},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "views_in_window", Value: -1}}}},
		{{Key: "$limit", Value: limit}},
		{{Key: "$lookup", Value: bson.M{
			"from":         "movies",
			"localField":   "_id",
			"foreignField": "_id",
			"as":           "movie_info",
		}}},
		{{Key: "$unwind", Value: bson.M{
			"path":                       "$movie_info",
			"preserveNullAndEmptyArrays": false,
		}}},
		{{Key: "$project", Value: bson.M{
			"_id":             "$movie_info._id",
			"title":           "$movie_info.title",
			"genre":           "$movie_info.genre",
			"thumbnail":       "$movie_info.thumbnail",
			"year":            "$movie_info.year",
			"views_in_window": 1,
			"total_views":     "$movie_info.view_count",
		}}},
	}

	cursor, err := r.historyCol.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []TrendingMovieResult
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

/**
 * 2. MOST WATCHED — All-time sort
 * MostWatchedResult is what GetMostWatched returns.
 * This is a simple sort — no aggregation needed.
 * view_count is incremented atomically every time someone streams a movie (Phase 3).
 * We rely on a MongoDB index on view_count (defined below in index hints).
*/
type MostWatchedResult struct {
	ID         primitive.ObjectID `bson:"_id"`
	Title      string             `bson:"title"`
	Genre      []string           `bson:"genre"`
	Thumbnail  string             `bson:"thumbnail"`
	Year       string             `bson:"year"`
	TotalViews int64              `bson:"view_count"`
}

/**
 * GetMostWatched returns movies sorted by all-time view_count.
 * This is a simple sort — no aggregation needed.
 * view_count is incremented atomically every time someone streams a movie (Phase 3).
 * We rely on a MongoDB index on view_count (defined below in index hints).
*/
func (r *AnalyticsRepository) GetMostWatched(ctx context.Context, limit int) ([]MostWatchedResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	opts := options.Find().
		SetSort(bson.D{{Key: "view_count", Value: -1}}).
		SetLimit(int64(limit)).
		SetProjection(bson.M{
			"title":      1,
			"genre":      1,
			"thumbnail":  1,
			"year":       1,
			"view_count": 1,
		})

	// Find movies sorted by view_count in descending order
	cursor, err := r.moviesCol.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []MostWatchedResult
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

/**
 * 3. TOP GENRES — Genre leaderboard
 * GenreStatResult is what GetTopGenres returns.
*/
type GenreStatResult struct {
	Genre      string `bson:"_id"`
	MovieCount int64  `bson:"movie_count"`
	TotalViews int64  `bson:"total_views"`
}

/**
 * GetTopGenres aggregates genre stats across all movies.
 * Pipeline:
 *  Stage 1 — $unwind:
 *    Each movie can have multiple genres (e.g. ["Horror","Sci-Fi"]).
 *    $unwind explodes one document into N documents — one per genre tag.
 *    So a movie with 3 genres becomes 3 separate documents in the pipeline.
 *
 *  Stage 2 — $group:
 *    Group by genre name.
 *    Count how many movies carry this genre (movie_count).
 *    Sum up view_count across all movies in this genre (total_views).
 *
 *  Stage 3 — $sort:
 *    Sort by total_views descending — most-viewed genre first.
 *
 *  Stage 4 — $limit:
 *    Trim to requested limit.
*/
func (r *AnalyticsRepository) GetTopGenres(ctx context.Context, limit int) ([]GenreStatResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 8*time.Second)
	defer cancel()

	pipeline := mongo.Pipeline{
		{{Key: "$unwind", Value: bson.M{
			"path":                       "$genre",
			"preserveNullAndEmptyArrays": false,
		}}},
		{{Key: "$group", Value: bson.M{
			"_id":         "$genre",
			"movie_count": bson.M{"$sum": 1},
			"total_views": bson.M{"$sum": "$view_count"},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "total_views", Value: -1}}}},
		{{Key: "$limit", Value: limit}},
	}

	cursor, err := r.moviesCol.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []GenreStatResult
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}

/**
 * 4. PLATFORM STATS — Collection-level counts
 * PlatformStatsResult is raw numbers from the DB.
 * This is used to get the overall statistics of the platform.
 * It is used in the admin dashboard.
*/
type PlatformStatsResult struct {
	TotalMovies    int64
	TotalUsers     int64
	TotalViews     int64
	TotalWatchlist int64
	TotalLikes     int64
}

/**
 * GetPlatformStats runs five parallel-safe counts + one aggregation.
 * We run them sequentially here for simplicity; in Phase 8 (production features)
 * these could be parallelised with goroutines + errgroup.
*/
func (r *AnalyticsRepository) GetPlatformStats(ctx context.Context) (*PlatformStatsResult, error) {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	// Count total movies
	totalMovies, err := r.moviesCol.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	// Count total users
	totalUsers, err := r.usersCol.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	// Sum all view_counts across all movies (single aggregation)
	// $group with _id: null collapses ALL documents into one and sums a field.
	totalViews, err := r.sumViewCounts(ctx)
	if err != nil {
		return nil, err
	}

	// Count total watchlist entries
	totalWatchlist, err := r.watchlistCol.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	/**
	 * FIX 1: r.likesCol  → r.reactionsCol  ("likes" collection never existed)
	 * FIX 2: "action":"like" → "type":"like"
	 * 
	 * Reaction model (internal/model/interaction.go):
	 *   Type ReactionType `bson:"type"`   ← the BSON field name is "type"
	 *   There is NO "action" field anywhere in the schema.
	 *   Filtering by "action" matched 0 documents — silent, no error.
	*/
	totalLikes, err := r.reactionsCol.CountDocuments(ctx, bson.M{"type": "like"})
	if err != nil {
		return nil, err
	}

	return &PlatformStatsResult{
		TotalMovies:    totalMovies,
		TotalUsers:     totalUsers,
		TotalViews:     totalViews,
		TotalWatchlist: totalWatchlist,
		TotalLikes:     totalLikes,
	}, nil
}

/**
 * sumViewCounts uses a $group aggregation to sum view_count across all movies.
 * This is more accurate than counting watch_history rows (view_count is the source of truth).
*/
func (r *AnalyticsRepository) sumViewCounts(ctx context.Context) (int64, error) {
	pipeline := mongo.Pipeline{
		{{Key: "$group", Value: bson.M{
			"_id":         nil,
			"total_views": bson.M{"$sum": "$view_count"},
		}}},
	}

	cursor, err := r.moviesCol.Aggregate(ctx, pipeline)
	if err != nil {
		return 0, err
	}
	defer cursor.Close(ctx)

	var result struct {
		TotalViews int64 `bson:"total_views"`
	}
	if cursor.Next(ctx) {
		if err := cursor.Decode(&result); err != nil {
			return 0, err
		}
	}
	return result.TotalViews, cursor.Err()
}

/**
 *  MongoDB Index Recommendations
 *
 *  EnsureAnalyticsIndexes creates the indexes needed for fast analytics queries.
 *  Call this once during app startup (in main.go or config/db.go).
 *
 *  Why these indexes?
 *
 *   watch_history.watched_at  → used by GetTrending's $match stage.
 *   Without this, Mongo does a full collection scan on every trending query.
 *
 *   movies.view_count         → used by GetMostWatched's sort.
 *   Without this, Mongo loads all movie docs just to sort them.
 *
 *   Since MongoDB 4.2+ all index builds are non-blocking by default,
 *   so SetBackground() is no longer needed.
*/
func EnsureAnalyticsIndexes(ctx context.Context, db *mongo.Database) error {
	// Index on watch_history.watched_at
	_, err := db.Collection("watch_history").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "events.watched_at", Value: -1}},
		Options: options.Index().SetName("idx_watch_history_watched_at"),
	})
	if err != nil {
		return err
	}

	// Index on movies.view_count
	_, err = db.Collection("movies").Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "view_count", Value: -1}},
		Options: options.Index().SetName("idx_movies_view_count"),
	})
	return err
}