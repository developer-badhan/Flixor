package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/developer-badhan/Flixor/internal/model"
)

/**
 * InteractionRepository defines all DB operations for user interactions.
 * By depending on an interface, we can easily mock this in tests.
*/
type InteractionRepository interface {
	// Watchlist
	AddToWatchlist(ctx context.Context, userID, movieID primitive.ObjectID) error
	RemoveFromWatchlist(ctx context.Context, userID, movieID primitive.ObjectID) error
	GetWatchlist(ctx context.Context, userID primitive.ObjectID) (*model.Watchlist, error)

	// Reactions
	UpsertReaction(ctx context.Context, userID, movieID primitive.ObjectID, reactionType model.ReactionType) error
	DeleteReaction(ctx context.Context, userID, movieID primitive.ObjectID) error
	GetReaction(ctx context.Context, userID, movieID primitive.ObjectID) (*model.Reaction, error)

	// Watch History
	AddToHistory(ctx context.Context, userID, movieID primitive.ObjectID) error
	GetHistory(ctx context.Context, userID primitive.ObjectID) (*model.WatchHistory, error)
	ClearHistory(ctx context.Context, userID primitive.ObjectID) error
}

/** 
 * interactionRepository is the concrete MongoDB implementation.
 * This struct holds the MongoDB collections for interactions.
 * watchlistCol: stores user watchlists
 * reactionCol: stores user reactions to movies
 * historyCol: stores user watch history	
*/
type interactionRepository struct {
	watchlistCol     *mongo.Collection
	reactionCol      *mongo.Collection
	historyCol    	 *mongo.Collection
	historyEventsCol *mongo.Collection
}

/**
 * NewInteractionRepository wires up the three collections and returns the repo.
 * Call this once at startup (in your DI/wire layer).
*/
func NewInteractionRepository(db *mongo.Database) InteractionRepository {
	repo := &interactionRepository{
		watchlistCol:     db.Collection("watchlists"),
		reactionCol:      db.Collection("reactions"),
		historyCol:       db.Collection("watch_history"),
		historyEventsCol: db.Collection("watch_events"),
	}

	// Ensure indexes so queries are fast and constraints are enforced at DB level.
	repo.ensureIndexes(context.Background())
	return repo
}

/**
 * ensureIndexes creates all required MongoDB indexes once at startup.
 * Indexes are idempotent — safe to run every time.
*/
func (r *interactionRepository) ensureIndexes(ctx context.Context) {
	// watchlists: one doc per user → unique index on user_id
	r.watchlistCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "user_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	// reactions: one reaction per (user, movie) pair → compound unique index
	r.reactionCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "user_id", Value: 1}, {Key: "movie_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	// watch_history: one doc per user → unique index on user_id
	r.historyCol.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys:    bson.D{{Key: "user_id", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
}

/**
 * AddToWatchlist pushes a movie into the user's watchlist array.
 * We use $addToSet logic via $push with $position — but to avoid real duplicates
 * we first check via the filter and use upsert + $addToSet equivalent.
 * Strategy: upsert the user doc, then $push the movie only if not already present
 * using the "movies.movie_id" filter trick.
*/
func (r *interactionRepository) AddToWatchlist(
	ctx context.Context,
	userID, movieID primitive.ObjectID,
) error {
	// Step 1: Upsert the watchlist document for this user if it doesn't exist yet.
	_, err := r.watchlistCol.UpdateOne(
		ctx,
		bson.M{
			"user_id":          userID,
			"movies.movie_id": bson.M{"$ne": movieID}, // Only add if not already present
		},
		bson.M{
			"$push": bson.M{
				"movies": model.WatchlistItem{
					MovieID: movieID,
					AddedAt: time.Now(),
				},
			},
			"$setOnInsert": bson.M{"user_id": userID},
		},
		options.Update().SetUpsert(true),
	)
	return err
}

/**
 * RemoveFromWatchlist pulls a movie out of the watchlist array using $pull.
 * This is used to remove a movie from the user's watchlist.
 * Example: 
 *   user_id: 123
 *   movie_id: 456
*/
func (r *interactionRepository) RemoveFromWatchlist(
	ctx context.Context,
	userID, movieID primitive.ObjectID,
) error {
	_, err := r.watchlistCol.UpdateOne(
		ctx,
		bson.M{"user_id": userID},
		bson.M{
			"$pull": bson.M{
				"movies": bson.M{"movie_id": movieID},
			},
		},
	)
	return err
}

/**
 * GetWatchlist returns the full watchlist document for a user.
 * Returns nil, nil if no watchlist exists yet (user hasn't added anything).
*/
func (r *interactionRepository) GetWatchlist(
	ctx context.Context,
	userID primitive.ObjectID,
) (*model.Watchlist, error) {
	var wl model.Watchlist
	err := r.watchlistCol.FindOne(ctx, bson.M{"user_id": userID}).Decode(&wl)
	if err == mongo.ErrNoDocuments {
		// Return an empty watchlist — not an error from the API's perspective.
		return &model.Watchlist{UserID: userID, Movies: []model.WatchlistItem{}}, nil
	}
	if err != nil {
		return nil, err
	}
	return &wl, nil
}

/**
 * UpsertReaction inserts a new reaction or updates an existing one.
 * Uses MongoDB's findOneAndUpdate with upsert=true so we never get duplicates.
*/
func (r *interactionRepository) UpsertReaction(
	ctx context.Context,
	userID, movieID primitive.ObjectID,
	reactionType model.ReactionType,
) error {
	filter := bson.M{"user_id": userID, "movie_id": movieID}
	update := bson.M{
		"$set": bson.M{
			"type":       reactionType,
			"updated_at": time.Now(),
		},
		"$setOnInsert": bson.M{
			"user_id":  userID,
			"movie_id": movieID,
		},
	}
	_, err := r.reactionCol.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))
	return err
}

/**
 * DeleteReaction removes the reaction document for a (user, movie) pair.
 * This is used to remove a reaction from a movie.
 * Example: 
 *   user_id: 123
 *   movie_id: 456
*/
func (r *interactionRepository) DeleteReaction(
	ctx context.Context,
	userID, movieID primitive.ObjectID,
) error {
	_, err := r.reactionCol.DeleteOne(ctx, bson.M{
		"user_id":  userID,
		"movie_id": movieID,
	})
	return err
}

/**
 * GetReaction fetches the reaction (like/dislike) for a specific (user, movie) pair.
 * Returns nil, nil if no reaction has been set yet.
*/
func (r *interactionRepository) GetReaction(
	ctx context.Context,
	userID, movieID primitive.ObjectID,
) (*model.Reaction, error) {
	var reaction model.Reaction
	err := r.reactionCol.FindOne(ctx, bson.M{
		"user_id":  userID,
		"movie_id": movieID,
	}).Decode(&reaction)

	if err == mongo.ErrNoDocuments {
		return nil, nil // No reaction yet is perfectly valid
	}
	if err != nil {
		return nil, err
	}
	return &reaction, nil
}

/**
* maxHistoryEvents is the cap on stored watch events per user.
* We slice from the end in the service layer to keep only the latest N entries.
*/
const maxHistoryEvents = 100

/**
 * AddToHistory prepends a new watch event to the user's history array.
 * We use $push with $each + $slice to automatically cap the array at maxHistoryEvents.
 * $slice: -100 keeps the LAST 100 items (most recent).
*/
func (r *interactionRepository) AddToHistory(
	ctx context.Context,
	userID, movieID primitive.ObjectID,
) error {
	newEvent := model.WatchEvent{
		MovieID:   movieID,
		WatchedAt: time.Now(),
	}

	_, err := r.historyCol.UpdateOne(
		ctx,
		bson.M{"user_id": userID},
		bson.M{
			"$push": bson.M{
				"events": bson.M{
					"$each":     []model.WatchEvent{newEvent},
					"$position": 0,                  // prepend (newest first)
					"$slice":    maxHistoryEvents,    // keep only latest N
				},
			},
			"$setOnInsert": bson.M{"user_id": userID},
		},
		options.Update().SetUpsert(true),
	)
	if err != nil {
		return err
	}
	// ALSO insert flat event for analytics
	_, err2 := r.historyEventsCol.InsertOne(ctx, bson.M{
		"user_id":    userID,
		"movie_id":   movieID,
		"watched_at": time.Now(),
	})
	if err2 != nil {
		return err2
	}
	return nil
}

/** 
 * GetHistory returns the watch history document for a user.
 * Events are stored newest-first due to $position: 0 in AddToHistory.
*/
func (r *interactionRepository) GetHistory(
	ctx context.Context,
	userID primitive.ObjectID,
) (*model.WatchHistory, error) {
	var history model.WatchHistory
	err := r.historyCol.FindOne(ctx, bson.M{"user_id": userID}).Decode(&history)
	if err == mongo.ErrNoDocuments {
		return &model.WatchHistory{UserID: userID, Events: []model.WatchEvent{}}, nil
	}
	if err != nil {
		return nil, err
	}
	return &history, nil
}

/**
 * ClearHistory deletes all watch events for a user by setting the events array to empty.
 * We don't delete the document itself to preserve the index entry.
*/
func (r *interactionRepository) ClearHistory(
	ctx context.Context,
	userID primitive.ObjectID,
) error {
	_, err := r.historyCol.UpdateOne(
		ctx,
		bson.M{"user_id": userID},
		bson.M{"$set": bson.M{"events": []model.WatchEvent{}}},
	)
	return err
}