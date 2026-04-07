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
	Identifier  string             `bson:"identifier"      json:"identifier"`  
	Title       string             `bson:"title"           json:"title"`
	Description string             `bson:"description"     json:"description"`
	Year        string             `bson:"year"            json:"year"`
	Genres      []string           `bson:"genres"          json:"genres"`      
	Director    string             `bson:"director"        json:"director"`    
	Subject     []string           `bson:"subject"         json:"subject"`
	ThumbnailURL string            `bson:"thumbnail_url"   json:"thumbnail_url"`
	StreamURL   string             `bson:"stream_url"      json:"stream_url"`
	ViewCount   int64              `bson:"view_count"      json:"view_count"`
	CreatedAt   time.Time          `bson:"created_at"      json:"created_at"`
	UpdatedAt   time.Time          `bson:"update_at"       json:"update_at"`
}

/**
 * MovieListResponse is a lightweight version used in list APIs
 * so we don't expose the full document unnecessarily.
*/
type MovieListResponse struct {
	ID           primitive.ObjectID `json:"id"`
	Title        string             `json:"title"`
	Genre        []string           `json:"genre"`
	Year         string             `json:"year"`
	Director     string             `json:"director"`
	ThumbnailURL string             `json:"thumbnail_url"`
	ViewCount    int64              `json:"view_count"`
}
