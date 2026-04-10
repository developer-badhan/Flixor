package model

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User is the single source of truth for the users collection in MongoDB.
type User struct {
	ID             primitive.ObjectID `bson:"_id,omitempty"      json:"id"`
	Username       string             `bson:"username"           json:"username"`
	Email          string             `bson:"email"              json:"email"`
	Password       string             `bson:"password"           json:"-"`
	ProfilePicture string             `bson:"profile_picture"    json:"profile_picture"`
	IsVerified     bool               `bson:"is_verified"        json:"is_verified"`
	OTPHash        string             `bson:"otp_hash"           json:"-"`
	OTPExpiresAt   time.Time          `bson:"otp_expires_at"     json:"-"`
	CreatedAt      time.Time          `bson:"created_at"         json:"created_at"`
	UpdatedAt      time.Time          `bson:"updated_at"         json:"updated_at"`
}

// PublicUser is the safe API projection of User.
type PublicUser struct {
	ID             primitive.ObjectID `json:"id"`
	Username       string             `json:"username"`
	Email          string             `json:"email"`
	ProfilePicture string             `json:"profile_picture"`
	IsVerified     bool               `json:"is_verified"`
	CreatedAt      time.Time          `json:"created_at"`
	UpdatedAt      time.Time          `json:"updated_at"`
}

// ToPublic converts a User to its safe API representation.
func (u *User) ToPublic() *PublicUser {
	return &PublicUser{
		ID:             u.ID,
		Username:       u.Username,
		Email:          u.Email,
		ProfilePicture: u.ProfilePicture,
		IsVerified:     u.IsVerified,
		CreatedAt:      u.CreatedAt,
		UpdatedAt:      u.UpdatedAt,
	}
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

// UpdateProfileRequest is the shape of the JSON body for PATCH /api/v1/user/profile.
type UpdateProfileRequest struct {
	Username string `json:"username" binding:"omitempty,min=3,max=32"`
	Password string `json:"password" binding:"omitempty,min=6"`
}

// VerifyOTPRequest is the body for POST /api/v1/user/verify-otp.
type VerifyOTPRequest struct {
	OTP string `json:"otp" binding:"required,len=6"`
}

// UpdateProfilePictureRequest is the body for PATCH /api/v1/user/profile/picture.
type UpdateProfilePictureRequest struct {
	ProfilePicture string `json:"profile_picture" binding:"required"`
}