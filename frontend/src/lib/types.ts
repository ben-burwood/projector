// Types mirroring the Go API structs.

export type SourceType =
  | "github"
  | "gitlab"
  | "gitea"
  | "dockerhub"
  | "pypi"
  | "crates"
  | "website";

export interface Tag {
  id: number;
  name: string;
  created_at: string;
  project_count?: number;
}

export interface Project {
  id: number;
  name: string;
  url: string;
  source_type: SourceType;
  source_identifier: string;
  source_host: string;
  source_meta: string;
  in_use: boolean;
  created_at: string;
  updated_at: string;
  tags: Tag[];
}

export interface Release {
  external_id: string;
  version: string;
  title: string;
  body: string;
  url: string;
  published_at: string; // RFC3339, may be empty
}

// A feed item pairs a release with the project it belongs to.
export interface FeedItem {
  project_id: number;
  project_name: string;
  source_type: SourceType;
  release: Release;
}

// Per-project error surfaced inline in the live feed.
export interface FeedError {
  project_id: number;
  project_name: string;
  error: string;
}

export interface FeedResponse {
  items: FeedItem[];
  errors: FeedError[];
  page: number;
  limit: number;
  has_more: boolean;
}

export interface DetectResponse {
  source_type: SourceType;
  source_identifier: string;
  source_host: string;
  prefill: {
    name: string;
    description: string;
    home_url: string;
  };
}

export interface ImportResult {
  total: number;
  imported: number;
  skipped: number;
}

export interface ProjectFilter {
  tag?: string;
  source?: string;
  q?: string;
  in_use?: boolean;
}
