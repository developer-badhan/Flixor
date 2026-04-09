package model

/**
 * TrendingQuery represents query parameters for GET /analytics/trending.
 * window: "1d" | "7d" | "30d"  (default: "7d")
 * limit:  1–100                 (default: 10)
*/
type TrendingQuery struct {
	Window string `form:"window" binding:"omitempty,oneof=1d 7d 30d"`
	Limit  int    `form:"limit"  binding:"omitempty,min=1,max=100"`
}

/**
 * TopQuery is a generic query param struct used by most-watched and top-genres.
 * limit:  1–100                 (default: 10)
*/
type TopQuery struct {
	Limit int `form:"limit" binding:"omitempty,min=1,max=100"`
}

/**
 * TrendingMovie is a single entry in the trending list.
 * ViewsInWindow is the count of views WITHIN the requested time window —
 * this is what makes it "trending", not the all-time view_count.
*/
type TrendingMovie struct {
	ID            string   `json:"id"`
	Title         string   `json:"title"`
	Genre         []string `json:"genre"`
	Thumbnail     string   `json:"thumbnail"`
	Year          string   `json:"year"`
	ViewsInWindow int64    `json:"views_in_window"` 
	TotalViews    int64    `json:"total_views"`    
	Rank          int      `json:"rank"`
}

/**
 * TrendingResponse wraps the trending list with metadata.
 * the window is the time window for which the movies are trending
 * the window_label is the label for the time window
 * the total_results is the total number of movies in the trending list
 * the movies is the list of trending movies
*/
type TrendingResponse struct {
	Window       string          `json:"window"`        
	WindowLabel  string          `json:"window_label"`  
	TotalResults int             `json:"total_results"`
	Movies       []TrendingMovie `json:"movies"`
}

/**
 * MostWatchedMovie is an entry in the all-time most-watched list.
 * the id is the movie id
 * the title is the movie title
 * the genre is the movie genre
 * the thumbnail is the movie thumbnail
 * the year is the movie year
 * the total_views is the total number of views for the movie
 * the rank is the rank of the movie in the all-time most-watched list
*/
type MostWatchedMovie struct {
	ID         string   `json:"id"`
	Title      string   `json:"title"`
	Genre      []string `json:"genre"`
	Thumbnail  string   `json:"thumbnail"`
	Year       string   `json:"year"`
	TotalViews int64    `json:"total_views"`
	Rank       int      `json:"rank"`
}

/**
 * MostWatchedResponse wraps the all-time list.
 * the total_results is the total number of movies in the all-time most-watched list
 * the movies is the list of all-time most-watched movies
*/
type MostWatchedResponse struct {
	TotalResults int                `json:"total_results"`
	Movies       []MostWatchedMovie `json:"movies"`
}

/**
 * GenreStat holds analytics for a single genre.
 * the genre is the genre name
 * the movie_count is the number of movies in the genre
 * the total_views is the total number of views for the genre
 * the rank is the rank of the genre in the genre leaderboard
*/
type GenreStat struct {
	Genre      string `json:"genre"`
	MovieCount int64  `json:"movie_count"` 
	TotalViews int64  `json:"total_views"` 
	Rank       int    `json:"rank"`
}

/**
 * TopGenresResponse wraps the genre leaderboard.
 * the total_results is the total number of genres in the genre leaderboard
 * the genres is the list of genres in the genre leaderboard
*/
type TopGenresResponse struct {
	TotalResults int         `json:"total_results"`
	Genres       []GenreStat `json:"genres"`
}

/**
 * PlatformStats is the high-level platform health snapshot.
 * the total_movies is the total number of movies on the platform
 * the total_users is the total number of users on the platform
 * the total_views is the total number of views on the platform
 * the total_watchlist is the total number of watchlist entries on the platform
 * the total_likes is the total number of likes on the platform
*/
type PlatformStats struct {
	TotalMovies    int64 `json:"total_movies"`
	TotalUsers     int64 `json:"total_users"`
	TotalViews     int64 `json:"total_views"`      
	TotalWatchlist int64 `json:"total_watchlist"` 
	TotalLikes     int64 `json:"total_likes"`
}