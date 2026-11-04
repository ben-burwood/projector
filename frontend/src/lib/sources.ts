import type { SourceType } from "./types";

interface SourceInfo {
  label: string;
  // Accent colour used for the source dot and project avatar tint.
  color: string;
  // Whether this source supports live release fetching.
  pollable: boolean;
}

export const SOURCES: Record<SourceType, SourceInfo> = {
  github: { label: "GitHub", color: "#8b949e", pollable: true },
  gitlab: { label: "GitLab", color: "#fc6d26", pollable: true },
  gitea: { label: "Gitea/Codeberg", color: "#609926", pollable: true },
  dockerhub: { label: "Docker Hub", color: "#1d9bf0", pollable: true },
  pypi: { label: "PyPI", color: "#7c6cf2", pollable: true },
  crates: { label: "crates.io", color: "#e0457b", pollable: true },
  website: { label: "Website", color: "#14b8a6", pollable: false },
};

export function sourceInfo(t: SourceType): SourceInfo {
  return SOURCES[t] ?? SOURCES.website;
}

export const SOURCE_TYPES = Object.keys(SOURCES) as SourceType[];
