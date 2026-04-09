package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

/**
 * RefreshToken represents a stored refresh token in MongoDB.
 * We store a SHA-256 hash of the actual token — never the raw value.
 *
 * Collection: refresh_tokens
 * Indexes recommended:
 *   - { token_hash: 1 }  unique
 *   - { user_id: 1 }
 *   - { expires_at: 1 }  TTL index (MongoDB auto-deletes expired docs)
*/
type RefreshToken struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"    json:"-"`
	UserID      primitive.ObjectID `bson:"user_id"          json:"-"`
	TokenHash   string             `bson:"token_hash"       json:"-"`
	Blacklisted bool               `bson:"blacklisted"      json:"-"`
	UserAgent   string             `bson:"user_agent"       json:"-"`
	IP          string             `bson:"ip"               json:"-"`
	ExpiresAt   time.Time          `bson:"expires_at"       json:"-"`
	CreatedAt   time.Time          `bson:"created_at"       json:"-"`
}

/**
 * TokenResponse is what we send back to the client after login or refresh.
 * Never expose internal IDs or hashes.
*/
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}