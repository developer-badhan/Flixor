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
 * RefreshTokenRepository defines the contract for refresh token persistence.
 * Coding to an interface means we can swap MongoDB for Redis or Postgres in tests.
*/
type RefreshTokenRepository interface {
	Create(ctx context.Context, token *model.RefreshToken) error
	FindByHash(ctx context.Context, hash string) (*model.RefreshToken, error)
	BlacklistByHash(ctx context.Context, hash string) error
	BlacklistAllForUser(ctx context.Context, userID primitive.ObjectID) error
	EnsureIndexes(ctx context.Context) error
}

/**
 * refreshTokenRepo implements RefreshTokenRepository using MongoDB.
 * It holds the "refresh_tokens" collection.
*/
type refreshTokenRepo struct {
	col *mongo.Collection
}

/**
 * NewRefreshTokenRepository wires the repo to the "refresh_tokens" collection.
 * This is the entry point for the repository.
 * It returns a RefreshTokenRepository.
*/
func NewRefreshTokenRepository(db *mongo.Database) RefreshTokenRepository {
	return &refreshTokenRepo{
		col: db.Collection("refresh_tokens"),
	}
}

/**
 * EnsureIndexes creates the necessary MongoDB indexes.
 * Call this once at application startup (in main.go or config).
 *
 * Indexes:
 *   - token_hash (unique)   — fast lookup + uniqueness guarantee
 *   - user_id               — fast "revoke all sessions for user"
 *   - expires_at (TTL=0)    — MongoDB auto-deletes expired tokens; no cron needed
*/
func (r *refreshTokenRepo) EnsureIndexes(ctx context.Context) error {
	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "token_hash", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "user_id", Value: 1}},
		},
		{   /**
			 * TTL index: MongoDB deletes the document when expires_at is reached.
			 * ExpireAfterSeconds: 0 means "delete at the exact expires_at time".
			*/
			Keys:    bson.D{{Key: "expires_at", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(0),
		},
	}

	_, err := r.col.Indexes().CreateMany(ctx, indexes)
	return err
}

/**
 * Create inserts a new refresh token document.
 * The token is created with the current time.
*/
func (r *refreshTokenRepo) Create(ctx context.Context, token *model.RefreshToken) error {
	token.CreatedAt = time.Now()
	_, err := r.col.InsertOne(ctx, token)
	return err
}

/**
 * FindByHash retrieves a token by its SHA-256 hash.
 * Returns mongo.ErrNoDocuments if not found — callers should handle this explicitly.
*/
func (r *refreshTokenRepo) FindByHash(ctx context.Context, hash string) (*model.RefreshToken, error) {
	var token model.RefreshToken
	filter := bson.M{"token_hash": hash}

	err := r.col.FindOne(ctx, filter).Decode(&token)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

/**
 * BlacklistByHash marks a single token as revoked.
 * We don't delete it so we can detect replay attacks in audit logs.
*/
func (r *refreshTokenRepo) BlacklistByHash(ctx context.Context, hash string) error {
	filter := bson.M{"token_hash": hash}
	update := bson.M{"$set": bson.M{"blacklisted": true}}

	_, err := r.col.UpdateOne(ctx, filter, update)
	return err
}

/**
 * BlacklistAllForUser revokes every refresh token for a user.
 * Use this on: password change, account compromise, "logout all devices".
*/
func (r *refreshTokenRepo) BlacklistAllForUser(ctx context.Context, userID primitive.ObjectID) error {
	filter := bson.M{"user_id": userID, "blacklisted": false}
	update := bson.M{"$set": bson.M{"blacklisted": true}}

	_, err := r.col.UpdateMany(ctx, filter, update)
	return err
}