package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

/**
 * Movie represents a movie document stored in MongoDB.
 * Fields map directly to what Internet Archive returns,
 * plus our own tracking fields (ViewCount, CreatedAt).
*/
type Movie struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"   json:"id"`
	Identifier  string             `bson:"identifier"      json:"identifier"`  // IA unique key, e.g. "CharlieChaplinTheKid"
	Title       string             `bson:"title"           json:"title"`
	Description string             `bson:"description"     json:"description"`
	Year        string             `bson:"year"            json:"year"`
	Genres      []string           `bson:"genres"          json:"genres"`      // "subject" field from IA
	Director    string             `bson:"director"        json:"director"`    // "creator" field from IA
	ThumbnailURL string            `bson:"thumbnail_url"   json:"thumbnail_url"`
	StreamURL   string             `bson:"stream_url"      json:"stream_url"`
	ViewCount   int64              `bson:"view_count"      json:"view_count"`
	CreatedAt   time.Time          `bson:"created_at"      json:"created_at"`
}
