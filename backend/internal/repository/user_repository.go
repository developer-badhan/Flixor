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
 * UserRepository handles all database operations for the users collection.
 * It owns the collection handle — nothing outside this file touches it directly.
*/
type UserRepository interface {
	CreateUser(ctx context.Context, user *model.User) (*model.User, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
	FindByID(ctx context.Context, id string) (*model.User, error)
	UpdateProfile(ctx context.Context, id string, updates bson.M) (*model.User, error)
	SaveOTP(ctx context.Context, id string, otpHash string, expiresAt time.Time) error
	ClearOTP(ctx context.Context, id string) error
	MarkVerified(ctx context.Context, id string) error
}

/**
 * userRepository implements UserRepository using MongoDB.
 * Unexported: callers receive the interface, never the concrete type.
*/
type userRepository struct {
	collection *mongo.Collection
}

/**
 * NewUserRepository creates a UserRepository wired to the users collection.
 * The service layer calls this once at startup and holds the result.
*/
func NewUserRepository(db *mongo.Database) UserRepository {
	return &userRepository{
		collection: db.Collection("users"),
	}
}

/**
 * CreateUser inserts a new user document into MongoDB.
 * Expects the password to already be hashed — this layer never touches plain text.
 * Sets CreatedAt and UpdatedAt server-side so the client cannot fake timestamps.
*/
func (r *userRepository) CreateUser(ctx context.Context, user *model.User) (*model.User, error) {
	now := time.Now()
	user.CreatedAt = now
	user.UpdatedAt = now
	user.ID = primitive.NewObjectID()

	result, err := r.collection.InsertOne(ctx, user)
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return nil, errors.New("email already registered")
		}
		return nil, errors.New("failed to create user")
	}

	user.ID = result.InsertedID.(primitive.ObjectID)
	return user, nil
}

/**
 * FindByEmail looks up a user by their email address.
 * Used during login to retrieve the stored password hash for verification.
 * Returns a typed ErrUserNotFound so the service can distinguish
 * "not found" from "database error" without string matching.
*/
func (r *userRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	var user model.User

	err := r.collection.FindOne(ctx, bson.M{"email": email}).Decode(&user)
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
func (r *userRepository) FindByID(ctx context.Context, id string) (*model.User, error) {
	var user model.User

	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	err = r.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, errors.New("failed to query user")
	}

	return &user, nil
}

/**
 * UpdateProfile updates a user's profile information.
 * It uses FindOneAndUpdate with SetReturnDocument(options.After) to return the updated document.
 * It also forces a server-side timestamp for the updated_at field.
*/
func (r *userRepository) UpdateProfile(ctx context.Context, id string, updates bson.M) (*model.User, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, errors.New("invalid user ID format")
	}

	// Force a server-side timestamp — the caller cannot override this.
	updates["updated_at"] = time.Now()

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var user model.User

	err = r.collection.FindOneAndUpdate(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": updates},
		opts,
	).Decode(&user)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, ErrUserNotFound
		}
		return nil, errors.New("failed to update user")
	}

	return &user, nil
}

/**
 * SaveOTP saves the OTP hash and expiry time for a user.
 * It uses UpdateOne with $set to update the otp_hash and otp_expires_at fields.
 * It also forces a server-side timestamp for the updated_at field.
*/
func (r *userRepository) SaveOTP(ctx context.Context, id string, otpHash string, expiresAt time.Time) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid user ID format")
	}

	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": bson.M{
			"otp_hash":       otpHash,
			"otp_expires_at": expiresAt,
			"updated_at":     time.Now(),
		}},
	)
	return err
}

/**
 * ClearOTP unsets the otp_hash and otp_expires_at fields.
 * Called after a failed verification attempt or explicit cleanup.
*/
func (r *userRepository) ClearOTP(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid user ID format")
	}

	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{"$unset": bson.M{"otp_hash": "", "otp_expires_at": ""}},
	)
	return err
}

/**
 * MarkVerified sets is_verified = true and atomically clears OTP fields.
 * Using a single UpdateOne ensures we never leave a user in a half-verified state.
*/
func (r *userRepository) MarkVerified(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return errors.New("invalid user ID format")
	}

	_, err = r.collection.UpdateOne(
		ctx,
		bson.M{"_id": objectID},
		bson.M{
			"$set":   bson.M{"is_verified": true, "updated_at": time.Now()},
			"$unset": bson.M{"otp_hash": "", "otp_expires_at": ""},
		},
	)
	return err
}
/**
 * ErrUserNotFound is returned when a user lookup finds no matching document.
 * The service layer checks for this to distinguish 404 from 500.
*/
var ErrUserNotFound = errors.New("user not found")