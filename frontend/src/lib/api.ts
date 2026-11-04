// Typed client for the Projector API. Same-origin: paths are relative so the
// Go server (which serves this SPA) answers them.

import type {
  DetectResponse,
  FeedResponse,
  ImportResult,
  Project,
  ProjectFilter,
  Release,
  Tag,
} from "./types";

export class ApiError extends Error {
  status: number;
  constructor(status: number, message: string) {
    super(message);
    this.status = status;
    this.name = "ApiError";
  }
}

async function request<T>(
  method: string,
  path: string,
  body?: unknown,
): Promise<T> {
  const opts: RequestInit = { method, headers: {} };
  if (body !== undefined) {
    (opts.headers as Record<string, string>)["Content-Type"] = "application/json";
    opts.body = JSON.stringify(body);
  }
  const res = await fetch(`/api${path}`, opts);
  if (res.status === 204) {
    return undefined as T;
  }
  const text = await res.text();
  const data = text ? JSON.parse(text) : undefined;
  if (!res.ok) {
    const msg = data?.error ?? `HTTP ${res.status}`;
    throw new ApiError(res.status, msg);
  }
  return data as T;
}

function query(params: object): string {
  const q = new URLSearchParams();
  for (const [k, v] of Object.entries(params)) {
    if (v !== undefined && v !== null && v !== "") q.set(k, String(v));
  }
  const s = q.toString();
  return s ? `?${s}` : "";
}

export const api = {
  meta: () => request<{ version: string; cache?: unknown }>("GET", "/meta"),

  // Projects
  listProjects: (f: ProjectFilter = {}) =>
    request<Project[]>("GET", `/projects${query(f)}`),
  getProject: (id: number) => request<Project>("GET", `/projects/${id}`),
  createProject: (p: Partial<Project> & { url: string; tag_ids?: number[] }) =>
    request<Project>("POST", "/projects", p),
  updateProject: (
    id: number,
    patch: Partial<Pick<Project, "name" | "url" | "in_use">> & {
      tag_ids?: number[];
    },
  ) => request<Project>("PATCH", `/projects/${id}`, patch),
  deleteProject: (id: number) => request<void>("DELETE", `/projects/${id}`),
  setProjectTags: (id: number, tagIds: number[]) =>
    request<Project>("PUT", `/projects/${id}/tags`, { tag_ids: tagIds }),

  // Live releases (Phase 3)
  projectReleases: (id: number, limit = 20) =>
    request<Release[]>("GET", `/projects/${id}/releases${query({ limit })}`),
  refreshProject: (id: number) =>
    request<void>("POST", `/projects/${id}/refresh`),

  // Tags
  listTags: () => request<Tag[]>("GET", "/tags"),
  createTag: (name: string) => request<Tag>("POST", "/tags", { name }),
  renameTag: (id: number, name: string) =>
    request<Tag>("PATCH", `/tags/${id}`, { name }),
  deleteTag: (id: number) => request<void>("DELETE", `/tags/${id}`),

  // Detection + feed (Phase 2/3)
  detect: (url: string) => request<DetectResponse>("POST", "/detect", { url }),
  feed: (
    f: { tag?: string; source?: string; all?: boolean; page?: number; limit?: number } = {},
  ) =>
    request<FeedResponse>("GET", `/feed${query(f)}`),
  refreshFeed: () => request<void>("POST", "/feed/refresh"),

  // Import
  importGitHubStars: (body: { username: string; in_use?: boolean; tag_ids?: number[] }) =>
    request<ImportResult>("POST", "/import/github-stars", body),
};
