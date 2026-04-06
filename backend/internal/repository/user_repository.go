package repository

import (
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/developer-badhan/Flixor/config"
	"github.com/developer-badhan/Flixor/internal/model"
)

/**
 * UserRepository handles all database operations for the users collection.
 * It owns the collection handle — nothing outside this file touches it directly.
*/
type UserRepository struct {
	collection *mongo.Collection
}

/**
 * NewUserRepository creates a UserRepository wired to the users collection.
 * The service layer calls this once at startup and holds the result.
*/
func NewUserRepository(db *config.DB) *UserRepository {
	return &UserRepository{
		collection: db.GetCollection("users"),
	}
}

/**
 * CreateUser inserts a new user document into MongoDB.
 * Expects the password to already be hashed — this layer never touches plain text.
 * Sets CreatedAt and UpdatedAt server-side so the client cannot fake timestamps.
*/
func (r *UserRepository) CreateUser(ctx context.Context, user *model.User) (*model.User, error) {
	// Always set timestamps server-side — never trust client-provided times
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now

	// Let MongoDB generate the ObjectID — guarantees uniqueness across all nodes
	user.ID = primitive.NewObjectID()

	result, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		// Check for duplicate key error — happens when email already exists.
		if mongo.IsDuplicateKeyError(err) {
			return nil, errors.New("email already registered")
		}
		return nil, errors.New("failed to create user")
	}

	// Confirm the inserted ID matches what we generated
	user.ID = result.InsertedID.(primitive.ObjectID)

	return user, nil
}

/**
 * FindByEmail looks up a user by their email address.
 * Used during login to retrieve the stored password hash for verification.
 * Returns a typed ErrUserNotFound so the service can distinguish
 * "not found" from "database error" without string matching.
*/
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User

	// bson.M is a map — {"email": email} becomes the MongoDB filter
	filter := bson.M{"email": email}

	err := r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, errors.New("failed to query user")
	}

	return &user, nil
}

/**
 * FindByID looks up a user by their MongoDB ObjectID.
 * Used by the auth middleware and future profile endpoints.
*/
func (r *UserRepository) FindByID(ctx context.Context, id string) (*model.User, error) {
	var user model.User

	/**
	 * Convert the string ID to MongoDB's ObjectID type before querying.
	 * If the string is not a valid ObjectID, reject it immediately.
	*/
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	// MongoDB stores the primary key as _id — bson.M{"_id": ...} maps to that
	filter := bson.M{"_id": objectID}

	err = r.collection.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, errors.New("failed to query user")
	}

	return &user, nil
}

/**
 * ErrUserNotFound is returned when a user lookup finds no matching document.
 * The service layer checks for this to distinguish 404 from 500.
*/
var ErrUserNotFound = errors.New("user not found")