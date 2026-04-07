package service

import (
	"context"
	"fmt"
	"sort"

	"github.com/developer-badhan/Flixor/internal/model"
	"github.com/developer-badhan/Flixor/internal/repository"
	
	"go.mongodb.org/mongo-driver/bson/primitive"
)

/**
 * Constants & Tuning Knobs
*/
const (
	defaultLimit     = 10 
	ruleEngineWeight = 1  
	aiEngineWeight   = 2  
)

/**
 * RecommendationService 
 * Workflow:
 *   - Build user profile (shared by both engines)
 *   - Run engine(s)
 *   - Sort by score, trim to limit
 *   - AI results score higher (more personalized signal)
*/
type RecommendationService struct {
	recoRepo     *repository.RecommendationRepository
	geminiClient *GeminiClient
}

func NewRecommendationService(
	recoRepo *repository.RecommendationRepository,
	geminiClient *GeminiClient,
) *RecommendationService {
	return &RecommendationService{
		recoRepo:     recoRepo,
		geminiClient: geminiClient,
	}
}

/**
 * GetRecommendations
 * This is used to get the recommendations.
 * It is used in the hybrid mode to get the recommendations.
*/
func (s *RecommendationService) GetRecommendations(
	ctx context.Context,
	userID primitive.ObjectID,
	req model.RecommendRequest,
) (*model.RecommendResponse, error) {

	limit := req.Limit
	if limit == 0 {
		limit = defaultLimit
	}

	mode := req.Mode
	if mode == "" {
		mode = "hybrid" // default: run both engines
	}

	// ── Step 1: Build user profile (shared by both engines) ──
	profile, err := s.buildUserProfile(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("recommendation: build profile: %w", err)
	}

	// scored results: key = movie ID hex string, value = scoredMovie
	scored := make(map[string]*scoredMovie)

	// ── Step 2: Run engine(s) ──
	switch mode {
	case "rule":
		ruleResults, err := s.runRuleEngine(ctx, profile, limit*2)
		if err != nil {
			return nil, fmt.Errorf("recommendation: rule engine: %w", err)
		}
		mergeIntoScored(scored, ruleResults, "rule", ruleEngineWeight)

	case "ai":
		aiResults, err := s.runAIEngine(ctx, profile, limit*2)
		if err != nil {
			return nil, fmt.Errorf("recommendation: ai engine: %w", err)
		}
		mergeAIIntoScored(scored, aiResults, aiEngineWeight)

	case "hybrid":
		// Run both; errors from one engine don't kill the whole request
		ruleResults, ruleErr := s.runRuleEngine(ctx, profile, limit*2)
		if ruleErr == nil {
			mergeIntoScored(scored, ruleResults, "rule", ruleEngineWeight)
		}

		aiResults, aiErr := s.runAIEngine(ctx, profile, limit*2)
		if aiErr == nil {
			mergeAIIntoScored(scored, aiResults, aiEngineWeight)
		}

		// If both engines failed, that's a real error
		if ruleErr != nil && aiErr != nil {
			return nil, fmt.Errorf("recommendation: both engines failed — rule: %v | ai: %v", ruleErr, aiErr)
		}
	}

	// ── Step 3: Sort by score, trim to limit ──
	ranked := rankAndTrim(scored, limit)

	return &model.RecommendResponse{
		Mode:         mode,
		TotalResults: len(ranked),
		Movies:       ranked,
	}, nil
}

/**
 * userProfile
 * This is used to store the user profile.
 * It is used in the hybrid mode to store the user profile.
*/
type userProfile struct {
	watchedIDs  []primitive.ObjectID 
	likedIDs    []primitive.ObjectID 
	topGenres   []string             
	genreFreq   map[string]int       
	watchTitles []string             
}

/**
 * buildUserProfile
 * This is used to build the user profile.
 * It is used in the hybrid mode to build the user profile.
*/
func (s *RecommendationService) buildUserProfile(ctx context.Context, userID primitive.ObjectID) (*userProfile, error) {
	// 1. Get watched movie IDs
	watchedIDs, err := s.recoRepo.GetWatchedMovieIDs(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 2. Get liked movie IDs
	likedIDs, err := s.recoRepo.GetLikedMovieIDs(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 3. Combine: watched + liked gives us the signal pool
	signalPool := append(watchedIDs, likedIDs...) 
	signalPool = append(signalPool, likedIDs...)   

	// 4. Get genre frequencies from the signal pool
	genreFreq, err := s.recoRepo.GetGenresFromMovieIDs(ctx, signalPool)
	if err != nil {
		return nil, err
	}

	// 5. Sort genres by frequency → strongest signal first
	topGenres := sortGenresByFrequency(genreFreq)

	// 6. Fetch titles of watched movies for Gemini prompt
	var watchTitles []string
	if len(watchedIDs) > 0 {
		watchMovies, err := s.recoRepo.GetMoviesByGenres(ctx, topGenres, nil, 5)
		if err == nil {
			for _, m := range watchMovies {
				watchTitles = append(watchTitles, m.Title)
			}
		}
	}

	return &userProfile{
		watchedIDs:  watchedIDs,
		likedIDs:    likedIDs,
		topGenres:   topGenres,
		genreFreq:   genreFreq,
		watchTitles: watchTitles,
	}, nil
}

/**
 * runRuleEngine
 * This is used to run the rule engine.
 * It is used in the hybrid mode to run the rule engine.
*/
func (s *RecommendationService) runRuleEngine(
	ctx context.Context,
	profile *userProfile,
	limit int,
) ([]ruleResult, error) {

	// Edge case: user has no genre signal yet → return empty (can't recommend)
	if len(profile.topGenres) == 0 {
		return nil, nil
	}

	// Take the top 3 genres for the query (too many genres = too broad)
	queryGenres := profile.topGenres
	if len(queryGenres) > 3 {
		queryGenres = queryGenres[:3]
	}

	movies, err := s.recoRepo.GetMoviesByGenres(ctx, queryGenres, profile.watchedIDs, limit)
	if err != nil {
		return nil, err
	}

	var results []ruleResult
	for _, m := range movies {
		reason := buildRuleReason(m.Genre, queryGenres, profile.genreFreq)
		results = append(results, ruleResult{movie: m, reason: reason})
	}
	return results, nil
}

/**
 * buildRuleReason
 * This is used to build the rule reason string.
 * It is used in the hybrid mode to build the rule reason string.
*/
func buildRuleReason(movieGenres, queryGenres []string, genreFreq map[string]int) string {
	var matched []string
	totalWatches := 0
	for _, mg := range movieGenres {
		for _, qg := range queryGenres {
			if mg == qg {
				matched = append(matched, mg)
				totalWatches += genreFreq[mg]
				break
			}
		}
	}
	if len(matched) == 0 {
		return "Popular in your preferred genres"
	}
	return fmt.Sprintf("Matches your top genres: %s (watched %d time(s) in these genres)", joinStrings(matched, ", "), totalWatches)
}

/**
 * runAIEngine
 * This is used to run the AI engine.
 * It is used in the hybrid mode to run the AI engine.
*/
func (s *RecommendationService) runAIEngine(
	ctx context.Context,
	profile *userProfile,
	limit int,
) ([]aiResult, error) {

	// Build the taste profile string Gemini will receive
	tasteProfile := buildTasteProfileString(profile)

	// Call Gemini — may return an error if API key is invalid / quota exceeded
	titles, err := s.geminiClient.RecommendMovies(ctx, tasteProfile)
	if err != nil {
		return nil, fmt.Errorf("gemini: %w", err)
	}

	if len(titles) == 0 {
		return nil, nil
	}

	// Fetch those movies from our DB (excludes already-watched)
	movies, err := s.recoRepo.GetMoviesByTitles(ctx, titles, profile.watchedIDs, limit)
	if err != nil {
		return nil, err
	}

	var results []aiResult
	for _, m := range movies {
		results = append(results, aiResult{
			movie:  m,
			reason: "Recommended by AI based on your taste profile",
		})
	}
	return results, nil
}

/**
 * buildTasteProfileString
 * This is used to build the taste profile string for the AI engine.
 * It is used in the hybrid mode to build the taste profile string for the AI engine.
*/
func buildTasteProfileString(profile *userProfile) string {
	if len(profile.topGenres) == 0 && len(profile.watchTitles) == 0 {
		return "The user is new with no watch history. Recommend popular classic films."
	}

	// Build the genre part of the taste profile string
	genrePart := ""
	if len(profile.topGenres) > 0 {
		top := profile.topGenres
		if len(top) > 5 {
			top = top[:5]
		}
		genrePart = fmt.Sprintf("Preferred genres (most watched first): %s.", joinStrings(top, ", "))
	}

	// Build the history part of the taste profile string
	historyPart := ""
	if len(profile.watchTitles) > 0 {
		historyPart = fmt.Sprintf("Recently watched movies: %s.", joinStrings(profile.watchTitles, ", "))
	}

	return fmt.Sprintf("%s %s", genrePart, historyPart)
}

/**
 * scoredMovie
 * This is used to store the scored movies.
 * It is used in the hybrid mode to store the scored movies.
*/
type scoredMovie struct {
	movie   repository.RecoMovie
	score   int
	sources []string
	reasons []string
}

/**
 * ruleResult
 * This is used to store the rule results.
 * It is used in the hybrid mode to store the rule results.
*/
type ruleResult struct {
	movie  repository.RecoMovie
	reason string
}

/**
 * aiResult
 * This is used to store the AI results.
 * It is used in the hybrid mode to store the AI results.
*/
type aiResult struct {
	movie  repository.RecoMovie
	reason string
}

// mergeIntoScored adds movies to the scored map.
// If the movie already exists (appeared in the other engine), its score increases.
// This naturally promotes movies that both engines agree on.
func mergeIntoScored(scored map[string]*scoredMovie, ruleResults []ruleResult, source string, weight int) {
	for _, r := range ruleResults {
		key := r.movie.ID.Hex()
		if existing, ok := scored[key]; ok {
			existing.score += weight
			existing.sources = append(existing.sources, source)
			existing.reasons = append(existing.reasons, r.reason)
		} else {
			scored[key] = &scoredMovie{
				movie:   r.movie,
				score:   weight,
				sources: []string{source},
				reasons: []string{r.reason},
			}
		}
	}
}

/**
 * mergeIntoScored for AI results (separate function for type safety)
 * This is used to merge the AI results into the scored map.
 * It is used in the hybrid mode to merge the AI results into the scored map.
*/
func mergeAIIntoScored(scored map[string]*scoredMovie, aiResults []aiResult, weight int) {
	for _, r := range aiResults {
		key := r.movie.ID.Hex()
		if existing, ok := scored[key]; ok {
			existing.score += weight
			existing.sources = append(existing.sources, "ai")
			existing.reasons = append(existing.reasons, r.reason)
		} else {
			scored[key] = &scoredMovie{
				movie:   r.movie,
				score:   weight,
				sources: []string{"ai"},
				reasons: []string{r.reason},
			}
		}
	}
}

/** 
 * rankAndTrim sorts movies by score (desc), then by view_count for tie-breaking,
 * trims to limit, and converts to the API response model.
*/
func rankAndTrim(scored map[string]*scoredMovie, limit int) []model.RecommendedMovie {
	list := make([]*scoredMovie, 0, len(scored))
	for _, sm := range scored {
		list = append(list, sm)
	}

	sort.Slice(list, func(i, j int) bool {
		if list[i].score != list[j].score {
			return list[i].score > list[j].score // higher score first
		}
		return list[i].movie.ViewCount > list[j].movie.ViewCount // tie-break by popularity
	})

	if len(list) > limit {
		list = list[:limit]
	}

	result := make([]model.RecommendedMovie, 0, len(list))
	for _, sm := range list {
		source := joinStrings(unique(sm.sources), "+")
		reason := sm.reasons[0] // show primary reason
		if len(sm.sources) > 1 {
			reason = "Top pick — matched by both genre analysis and AI"
		}

		result = append(result, model.RecommendedMovie{
			ID:          sm.movie.ID.Hex(),
			Title:       sm.movie.Title,
			Genre:       sm.movie.Genre,
			Thumbnail:   sm.movie.Thumbnail,
			Year:        sm.movie.Year,
			ViewCount:   sm.movie.ViewCount,
			Source:      source,
			MatchReason: reason,
		})
	}
	return result
}

// sortGenresByFrequency returns genres sorted by descending frequency.
func sortGenresByFrequency(freq map[string]int) []string {
	type kv struct {
		key string
		val int
	}
	var pairs []kv
	for k, v := range freq {
		pairs = append(pairs, kv{k, v})
	}
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].val > pairs[j].val
	})
	genres := make([]string, 0, len(pairs))
	for _, p := range pairs {
		genres = append(genres, p.key)
	}
	return genres
}

// joinStrings joins a slice of strings into a single string.
func joinStrings(ss []string, sep string) string {
	result := ""
	for i, s := range ss {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}

// unique returns a slice of unique strings.
func unique(ss []string) []string {
	seen := make(map[string]bool)
	var out []string
	for _, s := range ss {
		if !seen[s] {
			seen[s] = true
			out = append(out, s)
		}
	}
	return out
}