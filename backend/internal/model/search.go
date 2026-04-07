package model

/**
 * SearchFilter holds all query parameters for the movie search endpoint.
 * All fields are optional — the repository builds the MongoDB filter dynamically.
*/
type SearchFilter struct {
	Title string 
	Genre string 
	Page  int    
	Limit int    
}

/**
 * PaginatedMovies is the standard paginated response returned by GET /movies/search.
 * Frontend can use Total + Limit to calculate total pages.
*/
type PaginatedMovies struct {
	Total  int64   `json:"total"`  
	Page   int     `json:"page"`   
	Limit  int     `json:"limit"`  
	Pages  int64   `json:"pages"`  
	Movies []Movie `json:"movies"` 
}
