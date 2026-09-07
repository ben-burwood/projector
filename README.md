# projector

A self-hosted, single-user app to **save, organise and monitor the projects you like** from many sources — GitHub, GitLab, Codeberg/Gitea, Docker Hub, PyPI, crates.io, or any plain website.

It centralises the things you'd otherwise star on GitHub or follow on Docker Hub, lets you **tag** them and toggle **in-use**, and gives you an **aggregated feed** of the latest releases/versions across everything, plus a page per project.

Projector is a **live viewer, not a monitor**: it never stores release data and runs no background polling. Versions are fetched **live from each source's API on demand** (the project itself is the source of truth), with a short in-memory cache + ETag revalidation so reloads are fast and API-friendly. Only your bookmarks (projects + tags) are persisted, in SQLite.

### Inspiration
- https://newreleases.io/
- https://release-argus.io/

## Sources

| Source | Detect from | Live releases |
|---|---|---|
| GitHub | `github.com/owner/repo` | releases (falls back to tags) |
| GitLab | `gitlab.com/group/project` | releases |
| Codeberg / Gitea | `codeberg.org/owner/repo` (self-hosted via host) | releases |
| Docker Hub | `hub.docker.com/r/ns/img`, `/_/img` | image tags |
| PyPI | `pypi.org/project/name` | versions |
| crates.io | `crates.io/crates/name` | versions |
| Website | anything else | bookmark only |

Paste a URL on the **Add** page and Projector auto-detects the source and prefills the name.

## Tech

- **Backend:** Go (stdlib `net/http`), pure-Go SQLite (`modernc.org/sqlite`), single static binary.
- **Frontend:** Vue 3 + Vite + TypeScript + Tailwind v4 + daisyUI, embedded into the binary via `go:embed`.
- **Deploy:** multi-stage Docker → `FROM scratch`, published to GHCR on release.

## Run with Docker

```sh
docker compose up --build
```

Then open http://localhost:8080. Bookmarks persist in the `projector-data` volume.

Set a GitHub token to raise the API rate limit for live fetches:

```sh
PROJECTOR_GITHUB_TOKEN=ghp_xxx docker compose up --build
```

## Local development

The backend always serves the embedded SPA and stores its DB at `/data/projector.db`,
so run the whole app through Docker:

```sh
docker compose up --build
```

For fast frontend iteration, run the Vite dev server against the running backend:

```sh
cd frontend && npm install && npm run dev
```

## Configuration (env vars)

Everything is a compile-time constant except the optional provider tokens (used to
raise rate limits / reach private/self-hosted instances):

| Var | Default | Purpose |
|---|---|---|
| `PROJECTOR_GITHUB_TOKEN` | — | optional, raises GitHub rate limit |
| `PROJECTOR_GITLAB_TOKEN` | — | optional, private/self-hosted GitLab |
| `PROJECTOR_GITEA_TOKENS` | — | optional, `host=token,host=token` |

## API

`GET /api/meta` · `GET/POST /api/projects` · `GET/PATCH/DELETE /api/projects/{id}` ·
`GET /api/projects/{id}/releases` · `POST /api/projects/{id}/refresh` ·
`PUT/POST/DELETE /api/projects/{id}/tags` · `GET/POST/PATCH/DELETE /api/tags` ·
`POST /api/detect` · `GET /api/feed` · `POST /api/feed/refresh`
