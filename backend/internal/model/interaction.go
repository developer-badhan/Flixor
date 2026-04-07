package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

/**
 * WatchlistItem represents a single movie entry inside a user's watchlist.
 * One document per user — we upsert items into the Movies array.
 * The Movies array is capped at 100 items (enforced in service).
*/
type WatchlistItem struct {
	MovieID   primitive.ObjectID `bson:"movie_id"   json:"movie_id"`
	AddedAt   time.Time          `bson:"added_at"   json:"added_at"`
}

/**
 * Watchlist is a per-user document stored in the "watchlists" collection.
 * One document per user — we upsert items into the Movies array.
*/
type Watchlist struct {
	ID     primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID primitive.ObjectID `bson:"user_id"       json:"user_id"`
	Movies []WatchlistItem    `bson:"movies"        json:"movies"`
}


/**
 * ReactionType is an enum-style type for the two possible reactions.
 * The two possible reactions are "like" and "dislike".
*/
type ReactionType string
const (
	ReactionLike    ReactionType = "like"
	ReactionDislike ReactionType = "dislike"
)

/**
 * Reaction is one document per (user, movie) pair in the "reactions" collection.
 * A compound unique index on (user_id, movie_id) prevents duplicates at the DB level.
*/
type Reaction struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID   primitive.ObjectID `bson:"user_id"       json:"user_id"`
	MovieID  primitive.ObjectID `bson:"movie_id"      json:"movie_id"`
	Type     ReactionType       `bson:"type"          json:"type"`    // "like" | "dislike"
	UpdatedAt time.Time         `bson:"updated_at"    json:"updated_at"`
}


/**
 * WatchEvent represents a single instance of watching a movie.
 * We store it with a timestamp so we can show "last watched" info.
*/
type WatchEvent struct {
	MovieID   primitive.ObjectID `bson:"movie_id"   json:"movie_id"`
	WatchedAt time.Time          `bson:"watched_at" json:"watched_at"`
}

/**
 * WatchHistory is a per-user document in the "watch_history" collection.
 * We keep the latest 100 events per user (capped in the service layer).
*/
type WatchHistory struct {
	ID     primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID primitive.ObjectID `bson:"user_id"       json:"user_id"`
	Events []WatchEvent       `bson:"events"        json:"events"`
}

/**
 * ReactRequest is the JSON body expected by the react endpoint.
 * The Reaction field must be either "like" or "dislike".
*/
type ReactRequest struct {
	Reaction string `json:"reaction" binding:"required,oneof=like dislike"`
}