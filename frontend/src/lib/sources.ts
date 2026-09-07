import type { SourceType } from "./types";

interface SourceInfo {
  label: string;
  // DaisyUI badge modifier class for a bit of colour differentiation.
  badge: string;
  // Whether this source supports live release fetching.
  pollable: boolean;
}

export const SOURCES: Record<SourceType, SourceInfo> = {
  github: { label: "GitHub", badge: "badge-neutral", pollable: true },
  gitlab: { label: "GitLab", badge: "badge-warning", pollable: true },
  gitea: { label: "Gitea/Codeberg", badge: "badge-success", pollable: true },
  dockerhub: { label: "Docker Hub", badge: "badge-info", pollable: true },
  pypi: { label: "PyPI", badge: "badge-primary", pollable: true },
  crates: { label: "crates.io", badge: "badge-secondary", pollable: true },
  website: { label: "Website", badge: "badge-ghost", pollable: false },
};

export function sourceInfo(t: SourceType): SourceInfo {
  return SOURCES[t] ?? SOURCES.website;
}

export const SOURCE_TYPES = Object.keys(SOURCES) as SourceType[];
