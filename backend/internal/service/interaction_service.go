package service

import (
	"context"
	"errors"

	"github.com/developer-badhan/Flixor/internal/model"
	"github.com/developer-badhan/Flixor/internal/repository"
	"go.mongodb.org/mongo-driver/bson/primitive"
)


var (
	ErrInvalidReaction  = errors.New("reaction must be 'like' or 'dislike'")
	ErrNoReactionExists = errors.New("no reaction found for this movie")
)

/**
 * InteractionService defines all business operations.
 * Watchlist
 * Reactions
 * Watch History
*/
type InteractionService interface {
	// Watchlist
	AddToWatchlist(ctx context.Context, userID, movieID string) error
	RemoveFromWatchlist(ctx context.Context, userID, movieID string) error
	GetWatchlist(ctx context.Context, userID string) (*model.Watchlist, error)

	// Reactions
	ReactToMovie(ctx context.Context, userID, movieID, reactionType string) error
	RemoveReaction(ctx context.Context, userID, movieID string) error
	GetReaction(ctx context.Context, userID, movieID string) (*model.Reaction, error)

	// Watch History
	RecordWatch(ctx context.Context, userID, movieID string) error
	GetHistory(ctx context.Context, userID string) (*model.WatchHistory, error)
	ClearHistory(ctx context.Context, userID string) error
}

/** 
 * interactionService is the concrete implementation of InteractionService.
 * It holds a reference to the repository and implements the InteractionService interface.
 * The repo is injected at creation time.
*/
type interactionService struct {
	repo repository.InteractionRepository
}

/**
 * NewInteractionService creates the service and injects the repository dependency.
*/
func NewInteractionService(repo repository.InteractionRepository) InteractionService {
	return &interactionService{repo: repo}
}

/**
 * parseIDs converts string IDs to ObjectIDs and returns a clean error if invalid.
*/
func parseIDs(ids ...string) ([]primitive.ObjectID, error) {
	result := make([]primitive.ObjectID, 0, len(ids))
	for _, id := range ids {
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return nil, ErrInvalidMovieID
		}
		result = append(result, oid)
	}
	return result, nil
}

/**
 * AddToWatchlist validates IDs and delegates to the repository.
 * Business rule: adding a duplicate movie is a no-op (handled at DB layer via filter).
*/
func (s *interactionService) AddToWatchlist(ctx context.Context, userID, movieID string) error {
	ids, err := parseIDs(userID, movieID)
	if err != nil {
		return err
	}
	return s.repo.AddToWatchlist(ctx, ids[0], ids[1])
}

/**
 * RemoveFromWatchlist removes a movie from the watchlist.
 * Removing a movie that isn't in the list is a no-op — not an error.
*/
func (s *interactionService) RemoveFromWatchlist(ctx context.Context, userID, movieID string) error {
	ids, err := parseIDs(userID, movieID)
	if err != nil {
		return err
	}
	return s.repo.RemoveFromWatchlist(ctx, ids[0], ids[1])
}

/**
 * GetWatchlist fetches the user's watchlist.
 * Always returns a valid struct — even if empty (never nil on success).
*/
func (s *interactionService) GetWatchlist(ctx context.Context, userID string) (*model.Watchlist, error) {
	ids, err := parseIDs(userID)
	if err != nil {
		return nil, err
	}
	return s.repo.GetWatchlist(ctx, ids[0])
}

/**
 * ReactToMovie validates the reaction type and upserts the reaction.
 * Business rule: you can change from like → dislike (upsert handles this).
*/
func (s *interactionService) ReactToMovie(ctx context.Context, userID, movieID, reactionType string) error {
	// Validate reaction type
	rt := model.ReactionType(reactionType)
	if rt != model.ReactionLike && rt != model.ReactionDislike {
		return ErrInvalidReaction
	}

	ids, err := parseIDs(userID, movieID)
	if err != nil {
		return err
	}

	return s.repo.UpsertReaction(ctx, ids[0], ids[1], rt)
}

/**
 * RemoveReaction deletes the user's reaction for a movie.
 * If no reaction existed, we return a descriptive error so the API can return 404.
*/
func (s *interactionService) RemoveReaction(ctx context.Context, userID, movieID string) error {
	ids, err := parseIDs(userID, movieID)
	if err != nil {
		return err
	}

	// Check that a reaction actually exists before deleting
	existing, err := s.repo.GetReaction(ctx, ids[0], ids[1])
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrNoReactionExists
	}

	return s.repo.DeleteReaction(ctx, ids[0], ids[1])
}

/**
 * GetReaction fetches the current reaction for a (user, movie) pair.
 * Returns nil reaction + nil error if the user hasn't reacted yet.
*/
func (s *interactionService) GetReaction(ctx context.Context, userID, movieID string) (*model.Reaction, error) {
	ids, err := parseIDs(userID, movieID)
	if err != nil {
		return nil, err
	}
	return s.repo.GetReaction(ctx, ids[0], ids[1])
}

/**
 * RecordWatch logs a watch event for the user.
 * Business rule: we always record even if the same movie was watched before —
 * because watch history is time-based (each event has its own timestamp).
*/
func (s *interactionService) RecordWatch(ctx context.Context, userID, movieID string) error {
	ids, err := parseIDs(userID, movieID)
	if err != nil {
		return err
	}
	return s.repo.AddToHistory(ctx, ids[0], ids[1])
}

/**
 * GetHistory returns the user's watch history, newest events first.
 * Overrides the default behavior of returning the first 10 items.
*/
func (s *interactionService) GetHistory(ctx context.Context, userID string) (*model.WatchHistory, error) {
	ids, err := parseIDs(userID)
	if err != nil {
		return nil, err
	}
	return s.repo.GetHistory(ctx, ids[0])
}

/**
 * ClearHistory wipes all watch events for a user.
 * This is a destructive operation, so we should be careful when using it.
*/
func (s *interactionService) ClearHistory(ctx context.Context, userID string) error {
	ids, err := parseIDs(userID)
	if err != nil {
		return err
	}
	return s.repo.ClearHistory(ctx, ids[0])
}