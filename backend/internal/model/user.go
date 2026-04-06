package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

/**
 * User represents a registered user in the Flixor system.
 * This struct is the single source of truth for the users collection in MongoDB.
*/
type User struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"     json:"id"`
	Username  string             `bson:"username"          json:"username"`
	Email     string             `bson:"email"             json:"email"`
	Password  string             `bson:"password"          json:"-"`
	CreatedAt time.Time          `bson:"created_at"        json:"created_at"`
	UpdatedAt time.Time          `bson:"updated_at"        json:"updated_at"`
}

// RegisterRequest is the shape of the JSON body for POST /auth/register.
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=32"`
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginRequest is the shape of the JSON body for POST /auth/login.
type LoginRequest struct {
	Email    string `json:"email"    binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// AuthResponse is what we send back after a successful register or login.
type AuthResponse struct {
	Token string    `json:"token"`
	User  PublicUser `json:"user"`
}

/**
 * PublicUser is a safe view of the user — only fields we are happy
 * to expose over the API. The Password field from User never appears here.
*/
type PublicUser struct {
	ID        primitive.ObjectID `json:"id"`
	Username  string             `json:"username"`
	Email     string             `json:"email"`
	CreatedAt time.Time          `json:"created_at"`
}