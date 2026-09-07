package provider

import (
	"context"
	"sync"
	"time"
)

// maxFetch is the number of releases the service always fetches and caches per
// ref, regardless of the caller's requested limit. Keeping a single cache entry
// per ref (rather than one per requested limit) means the feed (small limit) and
// the project-detail page (larger limit) share one entry; callers just receive a
// prefix of the cached slice. Requests for more than this are capped.
const maxFetch = 50

// Service fetches releases through the registry, backed by an in-memory TTL
// cache with ETag revalidation. Nothing is persisted; the cache is cleared on
// restart. It is safe for concurrent use.
type Service struct {
	reg *Registry
	ttl time.Duration
	now func() time.Time // injectable for tests

	mu    sync.Mutex
	cache map[string]*entry
}

type entry struct {
	releases     []Release
	etag         string
	lastModified string
	fetchedAt    time.Time
	err          error // last error (served briefly to avoid hammering on failure)
}

// Stats summarises cache contents for /api/meta.
type Stats struct {
	Entries int `json:"entries"`
}

// NewService constructs a caching release service.
func NewService(reg *Registry, ttl time.Duration) *Service {
	return &Service{
		reg:   reg,
		ttl:   ttl,
		now:   time.Now,
		cache: make(map[string]*entry),
	}
}

func cacheKey(ref Ref) string {
	return ref.Type + "|" + ref.Host + "|" + ref.Identifier
}

// clampReturn bounds a caller's requested count.
func clampReturn(n int) int {
	if n <= 0 || n > maxFetch {
		return maxFetch
	}
	return n
}

// Releases returns up to `limit` releases for ref, serving from cache when fresh
// and revalidating with ETag on expiry. All access to cache entries happens
// under the mutex; the only unlocked work is the network fetch itself.
func (s *Service) Releases(ctx context.Context, ref Ref, limit int) ([]Release, error) {
	p := s.reg.Get(ref.Type)
	if p == nil {
		return nil, ErrNotPollable
	}

	key := cacheKey(ref)
	limit = clampReturn(limit)

	// Snapshot the current entry under the lock.
	s.mu.Lock()
	e, ok := s.cache[key]
	var (
		fresh        bool
		prevReleases []Release
		prevErr      error
		etag, lastMod string
	)
	if ok {
		fresh = s.now().Sub(e.fetchedAt) < s.ttl
		prevReleases = e.releases
		prevErr = e.err
		etag, lastMod = e.etag, e.lastModified
	}
	s.mu.Unlock()

	if fresh {
		if prevErr != nil {
			return nil, prevErr
		}
		return head(prevReleases, limit), nil
	}

	// Fetch outside the lock, using prior validators for a cheap 304.
	res, err := p.Fetch(ctx, FetchRequest{Ref: ref, ETag: etag, LastModified: lastMod, Limit: maxFetch})

	s.mu.Lock()
	defer s.mu.Unlock()

	if err != nil {
		// Cache the error briefly, preserving any prior validators.
		s.cache[key] = &entry{fetchedAt: s.now(), err: err, etag: etag, lastModified: lastMod}
		return nil, err
	}

	now := s.now()
	if res.NotModified {
		newEtag, newLM := etag, lastMod
		if res.ETag != "" {
			newEtag = res.ETag
		}
		if res.LastModified != "" {
			newLM = res.LastModified
		}
		// Rebuild the entry from the snapshot rather than mutating a pointer that
		// may have been replaced concurrently (e.g. by InvalidateAll).
		s.cache[key] = &entry{
			releases:     prevReleases,
			etag:         newEtag,
			lastModified: newLM,
			fetchedAt:    now,
		}
		return head(prevReleases, limit), nil
	}

	s.cache[key] = &entry{
		releases:     res.Releases,
		etag:         res.ETag,
		lastModified: res.LastModified,
		fetchedAt:    now,
	}
	return head(res.Releases, limit), nil
}

// head returns the first n elements of rs (or all of them if fewer).
func head(rs []Release, n int) []Release {
	if n >= len(rs) {
		return rs
	}
	return rs[:n]
}

// Invalidate drops the cache entry for ref so the next fetch is live.
func (s *Service) Invalidate(ref Ref) {
	s.mu.Lock()
	delete(s.cache, cacheKey(ref))
	s.mu.Unlock()
}

// InvalidateAll clears the entire cache.
func (s *Service) InvalidateAll() {
	s.mu.Lock()
	s.cache = make(map[string]*entry)
	s.mu.Unlock()
}

// Stats returns cache statistics.
func (s *Service) Stats() Stats {
	s.mu.Lock()
	defer s.mu.Unlock()
	return Stats{Entries: len(s.cache)}
}
