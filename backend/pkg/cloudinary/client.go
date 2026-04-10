package cloudinary

import (
	"context"
	"fmt"
	"io"

	"github.com/cloudinary/cloudinary-go/v2"
	"github.com/cloudinary/cloudinary-go/v2/api"
	"github.com/cloudinary/cloudinary-go/v2/api/uploader"
)

/**
 * Client wraps the Cloudinary SDK.
 * Constructed once at startup and injected into UserService.
*/
type Client struct {
	cld *cloudinary.Cloudinary
}

/**
 * NewClient initialises the Cloudinary SDK from config values.
 * Returns an error rather than calling log.Fatal — let main.go decide
 * whether a missing Cloudinary config is fatal.
*/
func NewClient(cloudName, apiKey, apiSecret string) (*Client, error) {
	cld, err := cloudinary.NewFromParams(cloudName, apiKey, apiSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to initialise cloudinary client: %w", err)
	}
	return &Client{cld: cld}, nil
}

/**
 * UploadProfilePicture uploads a profile image to the "flixor/profiles" folder.
 * Uses the userID as the PublicID so re-uploading overwrites the previous image —
 * no orphaned assets accumulate in your Cloudinary account.
 * Returns the HTTPS URL of the uploaded image.
*/
func (c *Client) UploadProfilePicture(ctx context.Context, file io.Reader, userID string) (string, error) {
	resp, err := c.cld.Upload.Upload(ctx, file, uploader.UploadParams{
		PublicID:     userID,
		Folder:       "flixor/profiles",
		Overwrite:    api.Bool(true),
		ResourceType: "image",
	})
	if err != nil {
		return "", fmt.Errorf("cloudinary upload failed: %w", err)
	}
	if resp.Error.Message != "" {
		return "", fmt.Errorf("cloudinary returned error: %s", resp.Error.Message)
	}

	return resp.SecureURL, nil
}