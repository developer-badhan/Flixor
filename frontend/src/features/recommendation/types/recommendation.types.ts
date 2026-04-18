// ─── Request ──────────────────────────────────────────────────────────────────

export type RecommendMode = 'rule' | 'ai' | 'hybrid';

export interface RecommendRequest {
  mode?: RecommendMode;   // default: "hybrid" (backend applies this default)
  limit?: number;         // 1–50, default: 10
}

// ─── Response — Normal ────────────────────────────────────────────────────────

/**
 * Source values the backend can return:
 *  "rule"     → rule engine only (genre-frequency matching)
 *  "ai"       → Gemini AI only
 *  "rule+ai"  → BOTH engines agreed → highest score → Top Pick
 */
export type RecommendSource = 'rule' | 'ai' | 'rule+ai';

export interface RecommendedMovie {
  id: string;
  title: string;
  genre: string[];
  thumbnail: string;      // analytics/reco model uses `thumbnail`, NOT `thumbnail_url`
  year: string;
  view_count: number;
  source: RecommendSource;
  match_reason: string;   // human-readable explanation — MUST be surfaced in UI
}

export interface RecommendResponse {
  mode: RecommendMode;
  total_results: number;
  movies: RecommendedMovie[];
}

// ─── Response — Empty State (new user, no history) ───────────────────────────
/**
 * Backend returns a DIFFERENT shape when no recommendations can be generated:
 * { message: string, movies: [] }
 * This has NO "success" key, so api.ts interceptor passes it through as-is.
 * We must detect this shape separately in the hook.
 */
export interface EmptyRecommendResponse {
  message: string;
  movies: never[];
}

// ─── Union type for hook consumers ───────────────────────────────────────────

export type AnyRecommendResponse = RecommendResponse | EmptyRecommendResponse;

// ─── Type guard ───────────────────────────────────────────────────────────────

export function isEmptyResponse(r: AnyRecommendResponse): r is EmptyRecommendResponse {
  return 'message' in r && (r as EmptyRecommendResponse).movies?.length === 0;
}