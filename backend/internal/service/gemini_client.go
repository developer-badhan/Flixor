package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

/**
 *  Gemini API types (request/response shapes)
 *  geminiRequest is the body we send to the Gemini REST API.
 *  geminiResponse is what the Gemini API returns.
 *  geminiContent is the content of the request.
 *  geminiPart is the part of the content.
 *  geminiGenerationConfig is the generation config of the request.
*/
type geminiRequest struct {
	Contents         []geminiContent        `json:"contents"`
	GenerationConfig geminiGenerationConfig `json:"generationConfig"`
}

/** 
 * geminiContent is the content of the request.
*/
type geminiContent struct {
	Parts []geminiPart `json:"parts"`
}

/**
 * geminiPart is the part of the content.
*/
type geminiPart struct {
	Text string `json:"text"`
}

/**
 * geminiGenerationConfig is the generation config of the request.
*/
type geminiGenerationConfig struct {
	Temperature     float64 `json:"temperature"`
	MaxOutputTokens int     `json:"maxOutputTokens"`
}

/**
 * geminiResponse is what the Gemini API returns.
*/
type geminiResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
	Error *struct {
		Message string `json:"message"`
		Code    int    `json:"code"`
	} `json:"error"`
}


// Gemini API URL
const geminiAPIURL = "https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent"

/**
 * GeminiClient wraps the Gemini REST API.
*/
type GeminiClient struct {
	apiKey     string
	httpClient *http.Client
}

/**
 * NewGeminiClient creates a new GeminiClient.
 * The HTTP client has a generous 15s timeout because Gemini can be slow under load.
*/
func NewGeminiClient(apiKey string) *GeminiClient {
	return &GeminiClient{
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

/**
 * RecommendMovies: the only public method we need
 * RecommendMovies sends the user's taste profile to Gemini and returns a list
 * of recommended movie titles. Caller maps these titles against the DB.
 * 
 * tasteProfile is a structured string describing what the user likes.
 * Example: "The user has watched: Metropolis, Nosferatu. Liked genres: Horror, Sci-Fi."
*/
func (g *GeminiClient) RecommendMovies(ctx context.Context, tasteProfile string) ([]string, error) {
	prompt := buildPrompt(tasteProfile)

	reqBody := geminiRequest{
		Contents: []geminiContent{
			{Parts: []geminiPart{{Text: prompt}}},
		},
		GenerationConfig: geminiGenerationConfig{
			Temperature:     0.7, // slight creativity, not chaotic
			MaxOutputTokens: 300,
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("gemini: marshal request: %w", err)
	}

	url := fmt.Sprintf("%s?key=%s", geminiAPIURL, g.apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("gemini: create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := g.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gemini: http request: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("gemini: read response: %w", err)
	}

	var geminiResp geminiResponse
	if err := json.Unmarshal(respBytes, &geminiResp); err != nil {
		return nil, fmt.Errorf("gemini: unmarshal response: %w", err)
	}

	// Surface API-level errors clearly
	if geminiResp.Error != nil {
		return nil, fmt.Errorf("gemini API error %d: %s", geminiResp.Error.Code, geminiResp.Error.Message)
	}

	if len(geminiResp.Candidates) == 0 {
		return nil, fmt.Errorf("gemini: no candidates in response")
	}

	rawText := geminiResp.Candidates[0].Content.Parts[0].Text
	titles := parseTitlesFromText(rawText)
	return titles, nil
}

/**
 * buildPrompt constructs the Gemini prompt.
 * We ask for a numbered list specifically — it's easy to parse reliably.
 * The prompt is a string that will be sent to the Gemini API.
 * The tasteProfile is a string that will be sent to the Gemini API.
 */
func buildPrompt(tasteProfile string) string {
	return fmt.Sprintf(`You are a movie recommendation assistant for Flixor, a streaming platform.

Based on this user's watch profile, recommend 10 classic or public domain movies they would enjoy.
The movies must be real films available in the public domain (pre-1928 or Creative Commons licensed).

User profile:
%s

Rules:
- Return ONLY a numbered list of movie titles, one per line.
- No descriptions, no explanations, no extra text.
- Format exactly like:
1. Movie Title One
2. Movie Title Two
3. Movie Title Three

Respond now:`, tasteProfile)
}

/**
 * parseTitlesFromText parses Gemini's numbered-list response into a clean string slice.
 * Input:  "1. Nosferatu\n2. The General\n3. Metropolis\n"
 * Output: ["Nosferatu", "The General", "Metropolis"]
*/
func parseTitlesFromText(raw string) []string {
	lines := strings.Split(raw, "\n")
	var titles []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Strip "1. " "12. " etc.
		if idx := strings.Index(line, ". "); idx != -1 && idx < 4 {
			line = strings.TrimSpace(line[idx+2:])
		}
		if line != "" {
			titles = append(titles, line)
		}
	}
	return titles
}