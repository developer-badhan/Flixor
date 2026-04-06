package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Internet Archive response shapes
/**
 * IASearchResponse is the top-level JSON envelope returned by IA's search API.
 * It contains a "response" field which holds the actual search results.
 * The "response" field has "numFound", "start", and "docs" which is a slice of IADoc.
 * Note: IA's API is a bit inconsistent, so we use json.RawMessage for fields that can be string or array.
*/
type IASearchResponse struct {
	Response IAResponseBody `json:"response"`
}

type IAResponseBody struct {
	NumFound int     `json:"numFound"`
	Start    int     `json:"start"`
	Docs     []IADoc `json:"docs"`
}

/**
 * IADoc is one movie entry from Internet Archive.
 * IA returns some fields as a string OR []string depending on the item,
 * so Subject and Creator use json.RawMessage for safe parsing.
*/
type IADoc struct {
	Identifier  string          `json:"identifier"`
	Title       string          `json:"title"`
	Description string          `json:"description"`
	// 
	Year        json.RawMessage `json:"year"`
	Subject     json.RawMessage `json:"subject"`  
	Creator     json.RawMessage `json:"creator"`  
}

// IAMovie is the cleaned, normalised version we hand back to the service.
type IAMovie struct {
	Identifier   string
	Title        string
	Description  string
	Year         string
	Genres       []string
	Director     string
	ThumbnailURL string
	StreamURL    string
}


// Public function
/**
 * FetchMoviesFromArchive calls the Internet Archive Advanced Search API and
 * returns a slice of normalised IAMovie structs.
 *
 * Parameters
 *    - rows : how many movies to fetch per request (max 100 recommended)
 *   - page : 1-based page number
*/
func FetchMoviesFromArchive(rows, page int) ([]IAMovie, error) {
	/**
	 * Build the request URL with required query parameters.
	 * fl[] tells IA which fields to include in the response.
	*/
	baseURL := "https://archive.org/advancedsearch.php"

	params := url.Values{}
	params.Set("q", "mediatype:movies AND licenseurl:(*creativecommons* OR *publicdomain*)")
	params.Add("fl[]", "identifier")
	params.Add("fl[]", "title")
	params.Add("fl[]", "description")
	params.Add("fl[]", "year")
	params.Add("fl[]", "subject")
	params.Add("fl[]", "creator")
	params.Set("rows", fmt.Sprintf("%d", rows))
	params.Set("page", fmt.Sprintf("%d", page))
	params.Set("output", "json")

	fullURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	// Use a client with a timeout — never use the default client in production.
	client := &http.Client{Timeout: 15 * time.Second}

	resp, err := client.Get(fullURL)
	if err != nil {
		return nil, fmt.Errorf("archive fetch: http get failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("archive fetch: unexpected status %d from %s", resp.StatusCode, fullURL)
	}

	var iaResp IASearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&iaResp); err != nil {
		return nil, fmt.Errorf("archive fetch: json decode failed: %w", err)
	}

	movies := make([]IAMovie, 0, len(iaResp.Response.Docs))
	for _, doc := range iaResp.Response.Docs {
		if strings.TrimSpace(doc.Identifier) == "" || strings.TrimSpace(doc.Title) == "" {
			continue // skip malformed entries
		}
		movies = append(movies, normalise(doc))
	}

	return movies, nil
}

// Helpers function
/** 
 * normalise takes an IADoc and returns a cleaned-up IAMovie. 
 * It trims whitespace, handles the string-or-array quirk for genres and director.
 * and constructs the thumbnail and stream URLs based on the identifier.
 * Note: IA's stream URL pattern is a best guess based on observed items and may not work for all entries.
*/
func normalise(doc IADoc) IAMovie {
	desc := strings.TrimSpace(doc.Description)

	// Optional safety: trim very long descriptions (frontend friendly)
	if len(desc) > 500 {
		desc = desc[:500]
	}

	return IAMovie{
		Identifier:   doc.Identifier,
		Title:        strings.TrimSpace(doc.Title),
		Description:  desc,
		Year:         parseYear(doc.Year),
		Genres:       parseStringOrSlice(doc.Subject),
		Director:     firstOf(parseStringOrSlice(doc.Creator)),
		ThumbnailURL: fmt.Sprintf("https://archive.org/services/img/%s", doc.Identifier),
		StreamURL:    fmt.Sprintf("https://archive.org/download/%s/%s.mp4", doc.Identifier, doc.Identifier),
	}
}

func parseYear(raw json.RawMessage) string {
	if len(raw) == 0 {
		return ""
	}

	// Try string
	var s string
	if err := json.Unmarshal(raw, &s); err == nil {
		return strings.TrimSpace(s)
	}

	// Try integer
	var n int
	if err := json.Unmarshal(raw, &n); err == nil {
		return fmt.Sprintf("%d", n)
	}

	// Try float (some APIs send 2014.0)
	var f float64
	if err := json.Unmarshal(raw, &f); err == nil {
		return fmt.Sprintf("%.0f", f)
	}

	return ""
}

/**
 * parseStringOrSlice handles the quirk where Internet Archive returns
 * a field as either a plain string or a JSON array of strings.
*/
func parseStringOrSlice(raw json.RawMessage) []string {
	if len(raw) == 0 {
		return []string{} 
	}

	// Try array first
	var arr []string
	if err := json.Unmarshal(raw, &arr); err == nil {
		return arr
	}

	// Fall back to plain string
	var s string
	if err := json.Unmarshal(raw, &s); err == nil && s != "" {
		return []string{s}
	}

	return []string{} 
}

// firstOf returns the first element of a slice, or empty string if empty.
func firstOf(s []string) string {
	if len(s) == 0 {
		return ""
	}
	return s[0]
}