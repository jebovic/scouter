# Product Images Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Download and store product images from option URLs in MinIO, expose them via API, and display them as a horizontal filmstrip in OptionCard.

**Architecture:** A new `internal/imagefetch` package provides a repository (DB), scraper (HTTP), uploader (MinIO), background worker (buffered channel, 2 goroutines), and nightly purge job. The research agent submits option IDs to the worker after each DB persist. Three REST endpoints expose images nested under the existing option route hierarchy. The frontend polls until images arrive and renders a horizontal filmstrip with a full-size lightbox.

**Tech Stack:** Go 1.25, MinIO Go client v7 (`github.com/minio/minio-go/v7`), pgx/v5, robfig/cron v3, React 19 + TypeScript, TanStack Query v5, CSS Modules, react-i18next.

---

## File Structure

### Backend — new files
| File | Responsibility |
|---|---|
| `backend/internal/db/migrations/023_option_images.up.sql` | Create `option_images` table |
| `backend/internal/db/migrations/023_option_images.down.sql` | Drop `option_images` table |
| `backend/internal/imagefetch/model.go` | `OptionImage` struct + `FetchedImage` struct |
| `backend/internal/imagefetch/repo.go` | DB CRUD for `option_images` (List, Insert, UpdateSortOrder, Delete) |
| `backend/internal/imagefetch/repo_test.go` | Integration tests against real DB |
| `backend/internal/imagefetch/scraper.go` | HTTP fetch + og:image/twitter/JSON-LD/img extraction |
| `backend/internal/imagefetch/scraper_test.go` | Unit tests with httptest.Server |
| `backend/internal/imagefetch/uploader.go` | MinIO client wrapper (upload, presign, bucket create) |
| `backend/internal/imagefetch/worker.go` | Buffered chan worker (cap 256, 2 goroutines) |
| `backend/internal/imagefetch/purge.go` | Nightly LRU quota purge via cron |
| `backend/internal/imagefetch/handler.go` | HTTP handlers: GET/PATCH/DELETE images |

### Backend — modified files
| File | Change |
|---|---|
| `backend/go.mod` / `go.sum` | Add `github.com/minio/minio-go/v7` |
| `backend/internal/research/agent.go` | Add `imageCh chan<- uuid.UUID` + `SetImageChannel()` + submit after persist |
| `backend/cmd/server/routes.go` | Add `imageWorker`, `imageHandler` to `routeDeps`; mount image routes |
| `backend/cmd/server/main.go` | Init MinIO, imageWorker, imageHandler; wire cron; wait on shutdown |
| `deployment/docker-compose.yml` | Add MinIO service + volume + Traefik labels |
| `.env.example` | Add MinIO env vars |

### Frontend — new files
| File | Responsibility |
|---|---|
| `frontend/src/api/images.ts` | Zod schema + `listOptionImages` fetch |
| `frontend/src/hooks/useOptionImages.ts` | TanStack Query hook with 30s poll + 50min staleTime |
| `frontend/src/components/options/ImageStrip.tsx` | Horizontal filmstrip (4 thumbnails + "+N" + skeleton) |
| `frontend/src/components/options/ImageStrip.module.css` | Filmstrip layout styles |
| `frontend/src/components/options/ImageLightbox.tsx` | Full-size modal lightbox |
| `frontend/src/components/options/ImageLightbox.module.css` | Lightbox overlay styles |

### Frontend — modified files
| File | Change |
|---|---|
| `frontend/src/components/options/OptionCard.tsx` | Import + render `<ImageStrip>` above card content |
| `frontend/src/i18n/en.json` | Add `options.images.loading`, `options.images.lightboxClose` |
| `frontend/src/i18n/fr.json` | Add French translations for image keys |

---

## Chunk 1: Database & Infrastructure

### Task 1: Migration 023 — option_images table

**Files:**
- Create: `backend/internal/db/migrations/023_option_images.up.sql`
- Create: `backend/internal/db/migrations/023_option_images.down.sql`

- [ ] **Step 1: Write the up migration**

```sql
-- backend/internal/db/migrations/023_option_images.up.sql
CREATE TABLE option_images (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    option_id    UUID NOT NULL REFERENCES options(id) ON DELETE CASCADE,
    minio_key    TEXT NOT NULL,
    content_type TEXT NOT NULL,
    width        INT,
    height       INT,
    source_url   TEXT,
    sort_order   INT NOT NULL DEFAULT 0,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX ON option_images(option_id);
```

- [ ] **Step 2: Write the down migration**

```sql
-- backend/internal/db/migrations/023_option_images.down.sql
DROP TABLE IF EXISTS option_images;
```

- [ ] **Step 3: Verify migration applies cleanly**

```bash
docker compose -f deployment/docker-compose.yml exec postgres \
  psql -U postgres -c '\dt option_images'
# Expected: table "option_images" listed
```

- [ ] **Step 4: Commit**

```bash
git add backend/internal/db/migrations/023_option_images.up.sql \
        backend/internal/db/migrations/023_option_images.down.sql
git commit -m "feat(db): migration 023 — option_images table"
```

---

### Task 2: MinIO infrastructure

**Files:**
- Modify: `deployment/docker-compose.yml`
- Modify: `.env.example`

- [ ] **Step 1: Add MinIO service to docker-compose.yml**

In `deployment/docker-compose.yml`, add under `services:`:

```yaml
  minio:
    image: minio/minio
    command: server /data
    environment:
      MINIO_ROOT_USER: ${MINIO_ACCESS_KEY}
      MINIO_ROOT_PASSWORD: ${MINIO_SECRET_KEY}
    volumes:
      - minio_data:/data
    networks:
      - scouter
    labels:
      - "traefik.enable=true"
      - "traefik.http.routers.minio.rule=Host(`minio.dev.local`)"
      - "traefik.http.routers.minio.entrypoints=websecure"
      - "traefik.http.routers.minio.tls=true"
      - "traefik.http.services.minio.loadbalancer.server.port=9000"
```

Also add `minio_data:` to the `volumes:` section at the bottom of the file.

> **Note:** `networks: [scouter]` must match the existing network name in `docker-compose.yml`. Read the file first and confirm the network name before adding.

- [ ] **Step 2: Add env vars to .env.example**

Append to `.env.example`:

```
# MinIO — object storage for product images
MINIO_ENDPOINT=minio:9000
MINIO_PUBLIC_URL=https://minio.dev.local
MINIO_ACCESS_KEY=scouter
MINIO_SECRET_KEY=scouter-secret
MINIO_BUCKET=scouter-images
MINIO_QUOTA_MB=2048
```

- [ ] **Step 3: Add MinIO env fields to config struct**

In `backend/internal/config/config.go`, add fields:

```go
MinioEndpoint  string `env:"MINIO_ENDPOINT"   envDefault:"minio:9000"`
MinioPublicURL string `env:"MINIO_PUBLIC_URL"  envDefault:"https://minio.dev.local"`
MinioAccessKey string `env:"MINIO_ACCESS_KEY"`
MinioSecretKey string `env:"MINIO_SECRET_KEY"`
MinioBucket    string `env:"MINIO_BUCKET"      envDefault:"scouter-images"`
MinioQuotaMB   int64  `env:"MINIO_QUOTA_MB"    envDefault:"2048"`
```

- [ ] **Step 4: Add MinIO dependency**

```bash
cd backend && go get github.com/minio/minio-go/v7@latest
go mod tidy
```

- [ ] **Step 5: Commit**

```bash
git add deployment/docker-compose.yml .env.example \
        backend/go.mod backend/go.sum \
        backend/internal/config/config.go
git commit -m "feat(infra): add MinIO service + env config"
```

---

## Chunk 2: imagefetch Package

### Task 3: Model types

**Files:**
- Create: `backend/internal/imagefetch/model.go`

- [ ] **Step 1: Write model.go**

```go
package imagefetch

import (
    "time"
    "github.com/google/uuid"
)

// OptionImage is a stored image record from the DB.
type OptionImage struct {
    ID          uuid.UUID `json:"id"`
    OptionID    uuid.UUID `json:"optionId"`
    MinioKey    string    `json:"-"`
    ContentType string    `json:"contentType"`
    Width       *int      `json:"width,omitempty"`
    Height      *int      `json:"height,omitempty"`
    SourceURL   string    `json:"-"`
    SortOrder   int       `json:"sortOrder"`
    CreatedAt   time.Time `json:"createdAt"`
    // URL is populated by the handler with a presigned URL — not stored in DB.
    URL string `json:"url,omitempty"`
}

// FetchedImage holds raw bytes from a successful scrape before MinIO upload.
type FetchedImage struct {
    Bytes       []byte
    ContentType string
    Width       *int
    Height      *int
    SourceURL   string
}
```

- [ ] **Step 2: Commit**

```bash
git add backend/internal/imagefetch/model.go
git commit -m "feat(imagefetch): model types"
```

---

### Task 4: Repository — DB read/write

**Files:**
- Create: `backend/internal/imagefetch/repo.go`
- Create: `backend/internal/imagefetch/repo_test.go`

- [ ] **Step 1: Write the failing test first**

```go
// backend/internal/imagefetch/repo_test.go
package imagefetch_test

import (
    "context"
    "testing"
    "github.com/google/uuid"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/jibei/scouter/internal/imagefetch"
    // testdb helper — reuse pattern from other _test files in the project
)

func TestRepo_InsertAndList(t *testing.T) {
    pool := testdb.Open(t)
    repo := imagefetch.NewRepository(pool)
    ctx := context.Background()

    optionID := uuid.New()
    // Insert a mission + option row so FK is satisfied (use test helpers)

    img, err := repo.Insert(ctx, imagefetch.InsertParams{
        OptionID:    optionID,
        MinioKey:    "options/abc/123.jpg",
        ContentType: "image/jpeg",
        SourceURL:   "https://example.com/img.jpg",
        SortOrder:   0,
    })
    require.NoError(t, err)
    assert.NotEqual(t, uuid.Nil, img.ID)
    assert.Equal(t, "options/abc/123.jpg", img.MinioKey)

    list, err := repo.ListByOption(ctx, optionID)
    require.NoError(t, err)
    assert.Len(t, list, 1)
    assert.Equal(t, img.ID, list[0].ID)
}

func TestRepo_SourceURLExists(t *testing.T) {
    pool := testdb.Open(t)
    repo := imagefetch.NewRepository(pool)
    ctx := context.Background()
    optionID := uuid.New()

    exists, err := repo.SourceURLExists(ctx, optionID, "https://example.com/img.jpg")
    require.NoError(t, err)
    assert.False(t, exists)
}

func TestRepo_Delete(t *testing.T) {
    pool := testdb.Open(t)
    repo := imagefetch.NewRepository(pool)
    ctx := context.Background()
    optionID := uuid.New()

    img, err := repo.Insert(ctx, imagefetch.InsertParams{
        OptionID:    optionID,
        MinioKey:    "options/test/del.jpg",
        ContentType: "image/jpeg",
        SourceURL:   "https://example.com/del.jpg",
        SortOrder:   0,
    })
    require.NoError(t, err)

    deleted, err := repo.Delete(ctx, img.ID)
    require.NoError(t, err)
    assert.Equal(t, img.ID, deleted.ID)

    list, err := repo.ListByOption(ctx, optionID)
    require.NoError(t, err)
    assert.Empty(t, list)
}
```

- [ ] **Step 2: Run the test to confirm it fails**

```bash
cd backend && go test ./internal/imagefetch/... 2>&1 | head -20
# Expected: compilation error — package does not exist yet
```

- [ ] **Step 3: Write repo.go**

```go
package imagefetch

import (
    "context"
    "fmt"
    "github.com/google/uuid"
    "github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct { pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) *Repository { return &Repository{pool: pool} }

type InsertParams struct {
    OptionID    uuid.UUID
    MinioKey    string
    ContentType string
    Width       *int
    Height      *int
    SourceURL   string
    SortOrder   int
}

func (r *Repository) Insert(ctx context.Context, p InsertParams) (*OptionImage, error) {
    row := r.pool.QueryRow(ctx, `
        INSERT INTO option_images (option_id, minio_key, content_type, width, height, source_url, sort_order)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id, option_id, minio_key, content_type, width, height, source_url, sort_order, created_at`,
        p.OptionID, p.MinioKey, p.ContentType, p.Width, p.Height, p.SourceURL, p.SortOrder,
    )
    return scanImage(row)
}

func (r *Repository) ListByOption(ctx context.Context, optionID uuid.UUID) ([]*OptionImage, error) {
    rows, err := r.pool.Query(ctx, `
        SELECT id, option_id, minio_key, content_type, width, height, source_url, sort_order, created_at
        FROM option_images WHERE option_id = $1 ORDER BY sort_order, created_at`,
        optionID,
    )
    if err != nil {
        return nil, fmt.Errorf("list images: %w", err)
    }
    defer rows.Close()
    var out []*OptionImage
    for rows.Next() {
        img, err := scanImage(rows)
        if err != nil {
            return nil, err
        }
        out = append(out, img)
    }
    return out, rows.Err()
}

func (r *Repository) SourceURLExists(ctx context.Context, optionID uuid.UUID, sourceURL string) (bool, error) {
    var exists bool
    err := r.pool.QueryRow(ctx,
        `SELECT EXISTS(SELECT 1 FROM option_images WHERE option_id=$1 AND source_url=$2)`,
        optionID, sourceURL,
    ).Scan(&exists)
    return exists, err
}

func (r *Repository) UpdateSortOrder(ctx context.Context, imageID uuid.UUID, sortOrder int) error {
    _, err := r.pool.Exec(ctx,
        `UPDATE option_images SET sort_order=$1 WHERE id=$2`,
        sortOrder, imageID,
    )
    return err
}

func (r *Repository) Delete(ctx context.Context, imageID uuid.UUID) (*OptionImage, error) {
    row := r.pool.QueryRow(ctx, `
        DELETE FROM option_images WHERE id=$1
        RETURNING id, option_id, minio_key, content_type, width, height, source_url, sort_order, created_at`,
        imageID,
    )
    return scanImage(row)
}

func (r *Repository) ListAll(ctx context.Context) ([]*OptionImage, error) {
    rows, err := r.pool.Query(ctx, `
        SELECT id, option_id, minio_key, content_type, width, height, source_url, sort_order, created_at
        FROM option_images ORDER BY created_at`)
    if err != nil {
        return nil, fmt.Errorf("list all images: %w", err)
    }
    defer rows.Close()
    var out []*OptionImage
    for rows.Next() {
        img, err := scanImage(rows)
        if err != nil {
            return nil, err
        }
        out = append(out, img)
    }
    return out, rows.Err()
}

type scanner interface{ Scan(dest ...any) error }

func scanImage(s scanner) (*OptionImage, error) {
    var img OptionImage
    err := s.Scan(
        &img.ID, &img.OptionID, &img.MinioKey, &img.ContentType,
        &img.Width, &img.Height, &img.SourceURL, &img.SortOrder, &img.CreatedAt,
    )
    if err != nil {
        return nil, fmt.Errorf("scan image: %w", err)
    }
    return &img, nil
}
```

- [ ] **Step 4: Run tests**

```bash
cd backend && go test ./internal/imagefetch/... -run TestRepo -v 2>&1
# Expected: PASS
```

- [ ] **Step 5: Commit**

```bash
git add backend/internal/imagefetch/
git commit -m "feat(imagefetch): repository — DB CRUD for option_images"
```

---

### Task 5: Scraper

**Files:**
- Create: `backend/internal/imagefetch/scraper.go`
- Create: `backend/internal/imagefetch/scraper_test.go`

- [ ] **Step 1: Write failing scraper test**

```go
// backend/internal/imagefetch/scraper_test.go
package imagefetch_test

import (
    "net/http"
    "net/http/httptest"
    "testing"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
    "github.com/jibei/scouter/internal/imagefetch"
)

func TestScraper_ExtractsOGImage(t *testing.T) {
    // Serve a fake page with og:image
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path == "/img.jpg" {
            w.Header().Set("Content-Type", "image/jpeg")
            w.Write(make([]byte, 1024)) // 1KB fake JPEG
            return
        }
        w.Header().Set("Content-Type", "text/html")
        fmt.Fprintf(w, `<html><head>
            <meta property="og:image" content="%s/img.jpg">
        </head><body></body></html>`, "http://"+r.Host)
    }))
    defer srv.Close()

    s := imagefetch.NewScraper()
    imgs, err := s.Fetch(t.Context(), srv.URL, nil)
    require.NoError(t, err)
    assert.Len(t, imgs, 1)
    assert.Equal(t, srv.URL+"/img.jpg", imgs[0].SourceURL)
    assert.Equal(t, "image/jpeg", imgs[0].ContentType)
}

func TestScraper_DeduplicatesSourceURL(t *testing.T) {
    // Page references same image via og:image and large <img> — scraper returns it only once.
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path == "/img.jpg" {
            w.Header().Set("Content-Type", "image/jpeg")
            w.Write(make([]byte, 1024))
            return
        }
        w.Header().Set("Content-Type", "text/html")
        imgURL := "http://" + r.Host + "/img.jpg"
        fmt.Fprintf(w, `<html><head>
            <meta property="og:image" content="%s">
        </head><body>
            <img src="%s" width="400">
        </body></html>`, imgURL, imgURL)
    }))
    defer srv.Close()

    s := imagefetch.NewScraper()
    imgs, err := s.Fetch(t.Context(), srv.URL, nil)
    require.NoError(t, err)
    assert.Len(t, imgs, 1, "duplicate URL should be deduplicated")
}

func TestScraper_SkipsSmallImages(t *testing.T) {
    // <img> without width ≥ 300 should be ignored
    srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if r.URL.Path == "/small.jpg" {
            w.Header().Set("Content-Type", "image/jpeg")
            w.Write(make([]byte, 1024))
            return
        }
        w.Header().Set("Content-Type", "text/html")
        imgURL := "http://" + r.Host + "/small.jpg"
        fmt.Fprintf(w, `<html><body>
            <img src="%s" width="100">
        </body></html>`, imgURL)
    }))
    defer srv.Close()

    s := imagefetch.NewScraper()
    imgs, err := s.Fetch(t.Context(), srv.URL, nil)
    require.NoError(t, err)
    assert.Empty(t, imgs, "small img (width<300) should be skipped")
}
```

- [ ] **Step 2: Run failing test**

```bash
cd backend && go test ./internal/imagefetch/... -run TestScraper -v 2>&1 | head -10
# Expected: FAIL — Scraper type not defined
```

- [ ] **Step 3: Write scraper.go**

```go
package imagefetch

import (
    "context"
    "fmt"
    "io"
    "log"
    "net/http"
    "net/url"
    "regexp"
    "strconv"
    "strings"
    "time"
)

const (
    maxImagesPerOption = 5
    maxImageBytes      = 5 * 1024 * 1024 // 5 MB
    scrapeTimeout      = 10 * time.Second
    minImgWidth        = 300
)

var userAgent = "Mozilla/5.0 (compatible; Scouter/1.0)"

type Scraper struct{ client *http.Client }

func NewScraper() *Scraper {
    return &Scraper{client: &http.Client{Timeout: scrapeTimeout}}
}

// Fetch scrapes pageURL and returns up to maxImagesPerOption images.
// existingURLs is the set of source_url values already stored for deduplication.
func (s *Scraper) Fetch(ctx context.Context, pageURL string, existingURLs map[string]bool) ([]FetchedImage, error) {
    req, _ := http.NewRequestWithContext(ctx, http.MethodGet, pageURL, nil)
    req.Header.Set("User-Agent", userAgent)

    resp, err := s.client.Do(req)
    if err != nil {
        return nil, fmt.Errorf("scrape %s: %w", pageURL, err)
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(io.LimitReader(resp.Body, 2*1024*1024))
    if err != nil {
        return nil, fmt.Errorf("read body: %w", err)
    }

    base, _ := url.Parse(pageURL)
    candidates := extractCandidates(string(body), base)

    var results []FetchedImage
    seen := make(map[string]bool)

    for _, candidate := range candidates {
        if len(results) >= maxImagesPerOption {
            break
        }
        if seen[candidate] || existingURLs[candidate] {
            continue
        }
        seen[candidate] = true

        img, err := s.downloadImage(ctx, candidate)
        if err != nil {
            log.Printf("imagefetch: skip %s: %v", candidate, err)
            continue
        }
        results = append(results, *img)
    }
    return results, nil
}

func (s *Scraper) downloadImage(ctx context.Context, imgURL string) (*FetchedImage, error) {
    req, _ := http.NewRequestWithContext(ctx, http.MethodGet, imgURL, nil)
    req.Header.Set("User-Agent", userAgent)
    resp, err := s.client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    data, err := io.ReadAll(io.LimitReader(resp.Body, maxImageBytes))
    if err != nil {
        return nil, err
    }
    if len(data) == 0 {
        return nil, fmt.Errorf("empty response")
    }

    ct := resp.Header.Get("Content-Type")
    if !strings.HasPrefix(ct, "image/") {
        ct = http.DetectContentType(data)
    }
    if !strings.HasPrefix(ct, "image/") {
        return nil, fmt.Errorf("not an image: %s", ct)
    }

    return &FetchedImage{
        Bytes:       data,
        ContentType: ct,
        SourceURL:   imgURL,
    }, nil
}

// extractCandidates returns image URLs in priority order:
// og:image, twitter:image, JSON-LD image, <img width≥300>
func extractCandidates(html string, base *url.URL) []string {
    var out []string
    seen := map[string]bool{}

    add := func(raw string) {
        u, err := base.Parse(strings.TrimSpace(raw))
        if err != nil || seen[u.String()] {
            return
        }
        seen[u.String()] = true
        out = append(out, u.String())
    }

    // og:image
    for _, m := range ogImageRE.FindAllStringSubmatch(html, -1) {
        add(m[1])
    }
    // twitter:image
    for _, m := range twitterImageRE.FindAllStringSubmatch(html, -1) {
        add(m[1])
    }
    // JSON-LD "image":
    for _, m := range jsonldImageRE.FindAllStringSubmatch(html, -1) {
        add(m[1])
    }
    // <img src="..." width="N"> where N >= 300
    for _, m := range imgTagRE.FindAllStringSubmatch(html, -1) {
        src, width := extractImgAttrs(m[0])
        if width >= minImgWidth {
            add(src)
        }
    }
    return out
}

var (
    ogImageRE      = regexp.MustCompile(`(?i)property=["']og:image["'][^>]*content=["']([^"']+)["']|content=["']([^"']+)["'][^>]*property=["']og:image["']`)
    twitterImageRE = regexp.MustCompile(`(?i)name=["']twitter:image["'][^>]*content=["']([^"']+)["']|content=["']([^"']+)["'][^>]*name=["']twitter:image["']`)
    jsonldImageRE  = regexp.MustCompile(`"image"\s*:\s*"([^"]+)"`)
    imgTagRE       = regexp.MustCompile(`(?i)<img\s[^>]+>`)
    imgSrcRE       = regexp.MustCompile(`(?i)src=["']([^"']+)["']`)
    imgWidthRE     = regexp.MustCompile(`(?i)width=["']?(\d+)["']?`)
)

func extractImgAttrs(tag string) (src string, width int) {
    if m := imgSrcRE.FindStringSubmatch(tag); m != nil {
        src = m[1]
    }
    if m := imgWidthRE.FindStringSubmatch(tag); m != nil {
        width, _ = strconv.Atoi(m[1])
    }
    return
}
```

- [ ] **Step 4: Run tests**

```bash
cd backend && go test ./internal/imagefetch/... -run TestScraper -v 2>&1
# Expected: PASS
```

- [ ] **Step 5: Commit**

```bash
git add backend/internal/imagefetch/scraper.go backend/internal/imagefetch/scraper_test.go
git commit -m "feat(imagefetch): HTML scraper — og:image, twitter, JSON-LD, img tags"
```

---

### Task 6: Uploader — MinIO client

**Files:**
- Create: `backend/internal/imagefetch/uploader.go`

- [ ] **Step 1: Write uploader.go**

```go
package imagefetch

import (
    "bytes"
    "context"
    "fmt"
    "net/url"
    "path/filepath"
    "sort"
    "strings"
    "time"

    "github.com/google/uuid"
    "github.com/minio/minio-go/v7"
    "github.com/minio/minio-go/v7/pkg/credentials"
)

const presignTTL = time.Hour

type UploaderConfig struct {
    Endpoint  string // internal Docker hostname, e.g. "minio:9000"
    PublicURL string // external URL for presigned URLs, e.g. "https://minio.dev.local"
    AccessKey string
    SecretKey string
    Bucket    string
}

type Uploader struct {
    client    *minio.Client
    bucket    string
    publicURL string
}

func NewUploader(ctx context.Context, cfg UploaderConfig) (*Uploader, error) {
    client, err := minio.New(cfg.Endpoint, &minio.Options{
        Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
        Secure: false, // TLS terminated by Traefik; plain HTTP on internal network
    })
    if err != nil {
        return nil, fmt.Errorf("minio client: %w", err)
    }

    exists, err := client.BucketExists(ctx, cfg.Bucket)
    if err != nil {
        return nil, fmt.Errorf("bucket check: %w", err)
    }
    if !exists {
        if err := client.MakeBucket(ctx, cfg.Bucket, minio.MakeBucketOptions{}); err != nil {
            return nil, fmt.Errorf("make bucket: %w", err)
        }
    }

    return &Uploader{client: client, bucket: cfg.Bucket, publicURL: cfg.PublicURL}, nil
}

// Upload stores img in MinIO and returns the object key.
func (u *Uploader) Upload(ctx context.Context, optionID uuid.UUID, img FetchedImage) (string, error) {
    ext := extensionForContentType(img.ContentType)
    key := fmt.Sprintf("options/%s/%s%s", optionID, uuid.New(), ext)

    _, err := u.client.PutObject(ctx, u.bucket, key,
        bytes.NewReader(img.Bytes), int64(len(img.Bytes)),
        minio.PutObjectOptions{ContentType: img.ContentType},
    )
    if err != nil {
        return "", fmt.Errorf("minio put %s: %w", key, err)
    }
    return key, nil
}

// PresignURL returns a time-limited URL for the given key, with the host
// rewritten to PublicURL so browsers receive HTTPS via Traefik.
func (u *Uploader) PresignURL(ctx context.Context, key string) (string, error) {
    raw, err := u.client.PresignedGetObject(ctx, u.bucket, key, presignTTL, url.Values{})
    if err != nil {
        return "", fmt.Errorf("presign %s: %w", key, err)
    }
    // Replace internal host with external public URL so browser gets HTTPS.
    pub, _ := url.Parse(u.publicURL)
    raw.Scheme = pub.Scheme
    raw.Host = pub.Host
    return raw.String(), nil
}

// Delete removes an object from MinIO.
func (u *Uploader) Delete(ctx context.Context, key string) error {
    return u.client.RemoveObject(ctx, u.bucket, key, minio.RemoveObjectOptions{})
}

// ListObjects returns all objects in the bucket sorted by LastModified (oldest first).
func (u *Uploader) ListObjects(ctx context.Context) ([]minio.ObjectInfo, error) {
    var objs []minio.ObjectInfo
    for obj := range u.client.ListObjects(ctx, u.bucket, minio.ListObjectsOptions{Recursive: true}) {
        if obj.Err != nil {
            return nil, obj.Err
        }
        objs = append(objs, obj)
    }
    // Sort oldest first for LRU eviction
    sort.Slice(objs, func(i, j int) bool {
        return objs[i].LastModified.Before(objs[j].LastModified)
    })
    return objs, nil
}

func extensionForContentType(ct string) string {
    switch strings.Split(ct, ";")[0] {
    case "image/jpeg":
        return ".jpg"
    case "image/png":
        return ".png"
    case "image/webp":
        return ".webp"
    case "image/gif":
        return ".gif"
    default:
        return filepath.Ext(ct)
    }
}
```

- [ ] **Step 2: Build check**

```bash
cd backend && go build ./internal/imagefetch/... 2>&1
# Expected: no errors
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/imagefetch/uploader.go
git commit -m "feat(imagefetch): MinIO uploader with presigned URL rewriting"
```

---

### Task 7: Worker

**Files:**
- Create: `backend/internal/imagefetch/worker.go`

- [ ] **Step 1: Write worker.go** (mirrors `internal/embedding/worker.go`)

```go
package imagefetch

import (
    "context"
    "log"
    "sync"

    "github.com/google/uuid"
)

const (
    channelCap  = 256
    workerCount = 2
)

// OptionJob carries the data needed to scrape and store images for one option.
// The research agent passes this struct to avoid importing the option package (circular dep).
type OptionJob struct {
    ID  uuid.UUID
    URL string
}

type Worker struct {
    jobs     chan OptionJob
    repo     *Repository
    uploader *Uploader
    scraper  *Scraper
    wg       sync.WaitGroup
}

func NewWorker(repo *Repository, uploader *Uploader, scraper *Scraper) *Worker {
    return &Worker{
        jobs:     make(chan OptionJob, channelCap),
        repo:     repo,
        uploader: uploader,
        scraper:  scraper,
    }
}

// Jobs returns the send-side of the job channel.
func (w *Worker) Jobs() chan<- OptionJob { return w.jobs }

// Submit enqueues a job for image fetching. Returns false if the channel is full (drop-on-full).
func (w *Worker) Submit(job OptionJob) bool {
    select {
    case w.jobs <- job:
        return true
    default:
        log.Printf("imagefetch: worker channel full, dropping %s", job.ID)
        return false
    }
}

// Start launches workerCount goroutines that process the job channel until ctx is done.
func (w *Worker) Start(ctx context.Context) {
    for range workerCount {
        w.wg.Add(1)
        go func() {
            defer w.wg.Done()
            for {
                select {
                case job, ok := <-w.jobs:
                    if !ok {
                        return
                    }
                    w.process(ctx, job)
                case <-ctx.Done():
                    return
                }
            }
        }()
    }
}

// Wait blocks until all goroutines have exited.
func (w *Worker) Wait() { w.wg.Wait() }

func (w *Worker) process(ctx context.Context, job OptionJob) {
    if job.URL == "" {
        return // no URL to scrape
    }
    existing, err := w.repo.ListByOption(ctx, job.ID)
    if err != nil {
        log.Printf("imagefetch: list existing for %s: %v", job.ID, err)
        return
    }
    existingURLs := make(map[string]bool, len(existing))
    for _, img := range existing {
        existingURLs[img.SourceURL] = true
    }

    imgs, err := w.scraper.Fetch(ctx, job.URL, existingURLs)
    if err != nil {
        log.Printf("imagefetch: scrape %s: %v", job.URL, err)
        return
    }

    for i, fi := range imgs {
        key, err := w.uploader.Upload(ctx, job.ID, fi)
        if err != nil {
            log.Printf("imagefetch: upload failed for %s: %v", job.ID, err)
            continue
        }
        _, err = w.repo.Insert(ctx, InsertParams{
            OptionID:    job.ID,
            MinioKey:    key,
            ContentType: fi.ContentType,
            Width:       fi.Width,
            Height:      fi.Height,
            SourceURL:   fi.SourceURL,
            SortOrder:   i,
        })
        if err != nil {
            log.Printf("imagefetch: insert image row for %s: %v", job.ID, err)
        }
    }
}
```

- [ ] **Step 2: Build check**

```bash
cd backend && go build ./internal/imagefetch/... 2>&1
# Expected: no errors
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/imagefetch/worker.go
git commit -m "feat(imagefetch): background worker — channel cap 256, 2 goroutines"
```

---

### Task 8: Purge job

**Files:**
- Create: `backend/internal/imagefetch/purge.go`

- [ ] **Step 1: Write purge.go**

```go
package imagefetch

import (
    "context"
    "log"

    "github.com/google/uuid"
)

// PurgeJob is a robfig/cron-compatible func that enforces the MinIO quota.
// quotaMB is the total allowed storage in mebibytes.
func PurgeJob(ctx context.Context, uploader *Uploader, repo *Repository, quotaMB int64) func() {
    quotaBytes := quotaMB * 1024 * 1024
    return func() {
        objs, err := uploader.ListObjects(ctx)
        if err != nil {
            log.Printf("imagefetch purge: list objects: %v", err)
            return
        }

        var total int64
        for _, o := range objs {
            total += o.Size
        }

        threshold := int64(float64(quotaBytes) * 0.90) // evict until below 90%
        if total <= quotaBytes {
            return // within quota
        }

        log.Printf("imagefetch purge: usage %d MB exceeds quota %d MB — evicting oldest objects",
            total/(1024*1024), quotaMB)

        for _, obj := range objs { // already sorted oldest-first by ListObjects
            if total <= threshold {
                break
            }
            if err := uploader.Delete(ctx, obj.Key); err != nil {
                log.Printf("imagefetch purge: delete object %s: %v", obj.Key, err)
                continue
            }
            if err := repo.DeleteByKey(ctx, obj.Key); err != nil {
                log.Printf("imagefetch purge: delete DB row for %s: %v", obj.Key, err)
            }
            total -= obj.Size
        }
    }
}
```

> Also add `DeleteByKey` to `repo.go`:

```go
// In repo.go — additional method:
func (r *Repository) DeleteByKey(ctx context.Context, minioKey string) error {
    _, err := r.pool.Exec(ctx, `DELETE FROM option_images WHERE minio_key=$1`, minioKey)
    return err
}
```

- [ ] **Step 2: Build check**

```bash
cd backend && go build ./internal/imagefetch/... 2>&1
# Expected: no errors
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/imagefetch/purge.go
git commit -m "feat(imagefetch): nightly LRU quota purge"
```

---

### Task 9: HTTP Handler

**Files:**
- Create: `backend/internal/imagefetch/handler.go`

- [ ] **Step 1: Write handler.go**

```go
package imagefetch

import (
    "encoding/json"
    "net/http"

    "github.com/go-chi/chi/v5"
    "github.com/google/uuid"
    "github.com/jibei/scouter/internal/httputil"
)

type Handler struct {
    repo     *Repository
    uploader *Uploader
}

func NewHandler(repo *Repository, uploader *Uploader) *Handler {
    return &Handler{repo: repo, uploader: uploader}
}

// Routes returns a chi router for image sub-routes.
// Mount at: r.Mount("/{optionID}/images", imageHandler.Routes())
func (h *Handler) Routes() http.Handler {
    r := chi.NewRouter()
    r.Get("/", h.list)
    r.Patch("/{imageID}", h.updateSortOrder)
    r.Delete("/{imageID}", h.delete)
    return r
}

func (h *Handler) list(w http.ResponseWriter, r *http.Request) {
    optionID, err := uuid.Parse(chi.URLParam(r, "optionID"))
    if err != nil {
        httputil.WriteError(w, http.StatusBadRequest, "invalid optionID")
        return
    }

    images, err := h.repo.ListByOption(r.Context(), optionID)
    if err != nil {
        httputil.WriteError(w, http.StatusInternalServerError, "list images")
        return
    }

    // Populate presigned URLs
    for _, img := range images {
        img.URL, _ = h.uploader.PresignURL(r.Context(), img.MinioKey)
    }
    httputil.WriteJSON(w, http.StatusOK, images)
}

type sortOrderBody struct {
    SortOrder int `json:"sortOrder"`
}

func (h *Handler) updateSortOrder(w http.ResponseWriter, r *http.Request) {
    imageID, err := uuid.Parse(chi.URLParam(r, "imageID"))
    if err != nil {
        httputil.WriteError(w, http.StatusBadRequest, "invalid imageID")
        return
    }
    var body sortOrderBody
    if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
        httputil.WriteError(w, http.StatusBadRequest, "invalid body")
        return
    }
    if err := h.repo.UpdateSortOrder(r.Context(), imageID, body.SortOrder); err != nil {
        httputil.WriteError(w, http.StatusInternalServerError, "update sort order")
        return
    }
    w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) delete(w http.ResponseWriter, r *http.Request) {
    imageID, err := uuid.Parse(chi.URLParam(r, "imageID"))
    if err != nil {
        httputil.WriteError(w, http.StatusBadRequest, "invalid imageID")
        return
    }

    img, err := h.repo.Delete(r.Context(), imageID)
    if err != nil {
        httputil.WriteError(w, http.StatusInternalServerError, "delete image")
        return
    }

    // Best-effort MinIO cleanup; log but don't fail.
    _ = h.uploader.Delete(r.Context(), img.MinioKey)

    w.WriteHeader(http.StatusNoContent)
}
```

- [ ] **Step 2: Build check**

```bash
cd backend && go build ./internal/imagefetch/... 2>&1
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/imagefetch/handler.go
git commit -m "feat(imagefetch): HTTP handler — GET/PATCH/DELETE image endpoints"
```

---

## Chunk 3: Server Wiring

### Task 10: Wire imagefetch into server

**Files:**
- Modify: `backend/cmd/server/main.go`
- Modify: `backend/internal/option/handler.go`
- Modify: `backend/internal/research/agent.go`

- [ ] **Step 1: Read routes.go and main.go** (confirm existing patterns for worker init and scheduler)

- [ ] **Step 2: Mount image routes inside option handler**

> `routeDeps` in routes.go does NOT need new fields — the image handler is wired directly in main.go via `WithImageHandler`.

- [ ] **Step 3: Mount image routes in registerRoutes**

In `registerRoutes`, find where optionHandler is mounted:
```go
r.Mount("/api/missions/{missionID}/options", d.optionHandler.Routes())
```

The image handler must nest inside the option routes. The cleanest approach is to mount it inside `optionHandler.Routes()`. Open `backend/internal/option/handler.go` and add inside the `Routes()` method:
```go
// At the end of Routes(), before return r:
if h.imageHandler != nil {
    r.Mount("/{optionID}/images", h.imageHandler.Routes())
}
```

Add `imageHandler *imagefetch.Handler` field to `option.Handler` struct and a setter:
```go
func (h *Handler) WithImageHandler(ih *imagefetch.Handler) {
    h.imageHandler = ih
}
```

- [ ] **Step 4: Initialize MinIO + workers in main.go**

After `embedWorker.Start(ctx)` (around line 105), add:
```go
// Image fetch worker
minioUploader, err := imagefetch.NewUploader(ctx, imagefetch.UploaderConfig{
    Endpoint:  cfg.MinioEndpoint,
    PublicURL: cfg.MinioPublicURL,
    AccessKey: cfg.MinioAccessKey,
    SecretKey: cfg.MinioSecretKey,
    Bucket:    cfg.MinioBucket,
})
if err != nil {
    log.Fatalf("minio init: %v", err)
}
imageRepo := imagefetch.NewRepository(pool)
imageScraper := imagefetch.NewScraper()
imageWorker := imagefetch.NewWorker(imageRepo, minioUploader, imageScraper)
imageWorker.Start(ctx)

researchAgent.SetImageChannel(imageWorker.Jobs())

imageHandler := imagefetch.NewHandler(imageRepo, minioUploader)
optionHandler.WithImageHandler(imageHandler)

// Nightly image quota purge — find the existing scheduler instance (sched from scheduler.New())
// that is already initialized in main.go and add the daily job to it.
sched.AddFunc("@daily", imagefetch.PurgeJob(ctx, minioUploader, imageRepo, cfg.MinioQuotaMB))
```

At shutdown (near `embedWorker.Wait()`), add:
```go
imageWorker.Wait()
```

- [ ] **Step 5: Add SetImageChannel to research agent**

In `backend/internal/research/agent.go`, mirror `SetEmbedChannel`:
```go
imageCh chan<- imagefetch.OptionJob

func (a *Agent) SetImageChannel(ch chan<- imagefetch.OptionJob) {
    a.imageCh = ch
}
```

After the line where embed channel is submitted (around line 142–145):
```go
if a.imageCh != nil {
    select {
    case a.imageCh <- imagefetch.OptionJob{ID: created.ID, URL: created.URL}:
    default:
        log.Printf("research: image worker channel full, dropping %s", created.ID)
    }
}
```

- [ ] **Step 6: Build the whole backend**

```bash
cd backend && go build ./... 2>&1
# Expected: no errors
```

- [ ] **Step 7: Run all backend tests**

```bash
cd backend && go test ./... 2>&1 | tail -20
# Expected: all PASS
```

- [ ] **Step 8: Commit**

```bash
git add backend/cmd/server/main.go \
        backend/internal/option/handler.go \
        backend/internal/research/agent.go
git commit -m "feat(imagefetch): wire worker + handler into server and research agent"
```

---

## Chunk 4: Frontend

### Task 11: API schema + fetch

**Files:**
- Create: `frontend/src/api/images.ts`

- [ ] **Step 1: Write images.ts**

```typescript
// frontend/src/api/images.ts
import { z } from "zod";

export const OptionImageSchema = z.object({
  id: z.string().uuid(),
  contentType: z.string(),
  width: z.number().optional(),
  height: z.number().optional(),
  sortOrder: z.number(),
  url: z.string().url(),
});

export type OptionImage = z.infer<typeof OptionImageSchema>;

export async function listOptionImages(
  missionId: string,
  optionId: string
): Promise<OptionImage[]> {
  const res = await fetch(
    `/api/missions/${missionId}/options/${optionId}/images`
  );
  if (!res.ok) throw new Error(`listOptionImages: ${res.status}`);
  const data = await res.json();
  return z.array(OptionImageSchema).parse(data);
}
```

- [ ] **Step 2: Type-check**

```bash
cd frontend && npx tsc --noEmit 2>&1 | head -20
# Expected: no errors
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/api/images.ts
git commit -m "feat(images): Zod schema + fetch for option images API"
```

---

### Task 12: useOptionImages hook

**Files:**
- Create: `frontend/src/hooks/useOptionImages.ts`

- [ ] **Step 1: Write the hook**

```typescript
// frontend/src/hooks/useOptionImages.ts
import { useQuery } from "@tanstack/react-query";
import { listOptionImages, type OptionImage } from "../api/images";

const POLL_INTERVAL_MS = 30_000;
const MAX_POLL_RETRIES = 10;
const STALE_TIME_MS = 50 * 60 * 1000; // 50 minutes — presigned URL TTL is 1h

export function useOptionImages(
  missionId: string,
  optionId: string
) {
  return useQuery<OptionImage[]>({
    queryKey: ["option-images", missionId, optionId],
    queryFn: () => listOptionImages(missionId, optionId),
    staleTime: STALE_TIME_MS,
    refetchOnWindowFocus: true,
    // Poll every 30s while no images found, up to 10 retries
    refetchInterval: (query) => {
      const data = query.state.data;
      const retries = query.state.dataUpdateCount;
      if (!data || data.length === 0 && retries < MAX_POLL_RETRIES) {
        return POLL_INTERVAL_MS;
      }
      return false;
    },
  });
}
```

- [ ] **Step 2: Type-check**

```bash
cd frontend && npx tsc --noEmit 2>&1 | head -20
```

- [ ] **Step 3: Commit**

```bash
git add frontend/src/hooks/useOptionImages.ts
git commit -m "feat(images): useOptionImages hook — 30s poll, 50min staleTime"
```

---

### Task 13: ImageStrip component

**Files:**
- Create: `frontend/src/components/options/ImageStrip.tsx`
- Create: `frontend/src/components/options/ImageStrip.module.css`

- [ ] **Step 1: Write the CSS module first**

```css
/* frontend/src/components/options/ImageStrip.module.css */
.strip {
  display: flex;
  gap: 8px;
  overflow-x: auto;
  padding-bottom: 4px;
  margin-bottom: 12px;
  scrollbar-width: none;
}

.strip::-webkit-scrollbar {
  display: none;
}

.thumb {
  min-width: 80px;
  width: 80px;
  height: 60px;
  object-fit: cover;
  border-radius: 8px;
  flex-shrink: 0;
  cursor: pointer;
  transition: opacity 0.15s;
}

.thumb:hover {
  opacity: 0.85;
}

.overflow {
  min-width: 80px;
  width: 80px;
  height: 60px;
  border-radius: 8px;
  flex-shrink: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  background: var(--surface-2, #1e293b);
  color: var(--text-dim, #94a3b8);
  font-size: 12px;
  font-weight: 500;
}

.skeleton {
  display: flex;
  gap: 8px;
  margin-bottom: 12px;
}

.skeletonThumb {
  min-width: 80px;
  width: 80px;
  height: 60px;
  border-radius: 8px;
  background: var(--surface-2, #1e293b);
  animation: pulse 1.5s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.4; }
}
```

- [ ] **Step 2: Write the component**

```tsx
// frontend/src/components/options/ImageStrip.tsx
import { useState } from "react";
import { useTranslation } from "react-i18next";
import { type OptionImage } from "../../api/images";
import { ImageLightbox } from "./ImageLightbox";
import styles from "./ImageStrip.module.css";

const MAX_VISIBLE = 4;

interface Props {
  images: OptionImage[];
  loading: boolean;
}

export function ImageStrip({ images, loading }: Props) {
  const { t } = useTranslation();
  const [lightboxIndex, setLightboxIndex] = useState<number | null>(null);

  if (loading && images.length === 0) {
    return (
      <div className={styles.skeleton}>
        {[0, 1, 2].map((i) => (
          <div key={i} className={styles.skeletonThumb} />
        ))}
      </div>
    );
  }

  if (images.length === 0) return null;

  const visible = images.slice(0, MAX_VISIBLE);
  const overflow = images.length - MAX_VISIBLE;

  return (
    <>
      <div className={styles.strip} role="list" aria-label={t("options.images.gallery")}>
        {visible.map((img, i) => (
          <img
            key={img.id}
            src={img.url}
            alt=""
            className={styles.thumb}
            width={80}
            height={60}
            onClick={() => setLightboxIndex(i)}
          />
        ))}
        {overflow > 0 && (
          <div
            className={styles.overflow}
            onClick={() => setLightboxIndex(MAX_VISIBLE - 1)}
          >
            +{overflow}
          </div>
        )}
      </div>
      {lightboxIndex !== null && (
        <ImageLightbox
          images={images}
          initialIndex={lightboxIndex}
          onClose={() => setLightboxIndex(null)}
        />
      )}
    </>
  );
}
```

- [ ] **Step 3: Type-check**

```bash
cd frontend && npx tsc --noEmit 2>&1 | head -20
```

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/options/ImageStrip.tsx \
        frontend/src/components/options/ImageStrip.module.css
git commit -m "feat(images): ImageStrip component — filmstrip, skeleton, overflow chip"
```

---

### Task 14: ImageLightbox component

**Files:**
- Create: `frontend/src/components/options/ImageLightbox.tsx`
- Create: `frontend/src/components/options/ImageLightbox.module.css`

- [ ] **Step 1: Write the CSS**

```css
/* frontend/src/components/options/ImageLightbox.module.css */
.overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.85);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.image {
  max-width: 90vw;
  max-height: 85vh;
  object-fit: contain;
  border-radius: 8px;
}

.close {
  position: absolute;
  top: 16px;
  right: 16px;
  background: none;
  border: none;
  color: #fff;
  font-size: 28px;
  cursor: pointer;
  line-height: 1;
  padding: 4px 8px;
}

.close:hover {
  opacity: 0.7;
}

.nav {
  position: absolute;
  top: 50%;
  transform: translateY(-50%);
  background: rgba(0,0,0,0.5);
  border: none;
  color: #fff;
  font-size: 24px;
  cursor: pointer;
  padding: 12px 16px;
  border-radius: 4px;
}

.nav:hover { background: rgba(0,0,0,0.8); }

.navPrev { left: 16px; }
.navNext { right: 16px; }
```

- [ ] **Step 2: Write the component**

```tsx
// frontend/src/components/options/ImageLightbox.tsx
import { useState, useEffect } from "react";
import { useTranslation } from "react-i18next";
import { type OptionImage } from "../../api/images";
import styles from "./ImageLightbox.module.css";

interface Props {
  images: OptionImage[];
  initialIndex: number;
  onClose: () => void;
}

export function ImageLightbox({ images, initialIndex, onClose }: Props) {
  const { t } = useTranslation();
  const [index, setIndex] = useState(initialIndex);

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
      if (e.key === "ArrowLeft") setIndex((i) => Math.max(0, i - 1));
      if (e.key === "ArrowRight") setIndex((i) => Math.min(images.length - 1, i + 1));
    };
    window.addEventListener("keydown", onKey);
    return () => window.removeEventListener("keydown", onKey);
  }, [images.length, onClose]);

  const img = images[index];

  return (
    <div className={styles.overlay} onClick={onClose}>
      <button
        className={styles.close}
        onClick={onClose}
        aria-label={t("options.images.lightboxClose")}
      >
        ×
      </button>
      <img
        src={img.url}
        alt=""
        className={styles.image}
        onClick={(e) => e.stopPropagation()}
      />
      {images.length > 1 && (
        <>
          <button
            className={`${styles.nav} ${styles.navPrev}`}
            onClick={(e) => { e.stopPropagation(); setIndex((i) => Math.max(0, i - 1)); }}
            disabled={index === 0}
          >
            ‹
          </button>
          <button
            className={`${styles.nav} ${styles.navNext}`}
            onClick={(e) => { e.stopPropagation(); setIndex((i) => Math.min(images.length - 1, i + 1)); }}
            disabled={index === images.length - 1}
          >
            ›
          </button>
        </>
      )}
    </div>
  );
}
```

- [ ] **Step 3: Type-check**

```bash
cd frontend && npx tsc --noEmit 2>&1 | head -20
```

- [ ] **Step 4: Commit**

```bash
git add frontend/src/components/options/ImageLightbox.tsx \
        frontend/src/components/options/ImageLightbox.module.css
git commit -m "feat(images): ImageLightbox component — keyboard navigation, overlay close"
```

---

### Task 15: Wire into OptionCard + i18n keys

**Files:**
- Modify: `frontend/src/components/options/OptionCard.tsx`
- Modify: `frontend/src/pages/OptionsExplorer.tsx`
- Modify: `frontend/src/i18n/en.json`
- Modify: `frontend/src/i18n/fr.json`
- Modify: `frontend/src/i18n/de.json`

- [ ] **Step 1: Read OptionCard.tsx to find injection point**

```bash
head -60 frontend/src/components/options/OptionCard.tsx
```

- [ ] **Step 2: Check OptionsExplorer for missionId threading**

Read `frontend/src/pages/OptionsExplorer.tsx` and verify that `missionId` is passed to `OptionCard`. If `OptionCard` doesn't yet accept `missionId` as a prop, add it now. Also read how `OptionsExplorer` gets `missionId` (likely from `useParams()` or `useMission()`). Confirm the prop flows through before wiring the image hook.

- [ ] **Step 3: Add ImageStrip to OptionCard**

Import and add above the card content:

```tsx
import { ImageStrip } from "./ImageStrip";
import { useOptionImages } from "../../hooks/useOptionImages";

// Inside the component, after props destructuring:
const { data: images = [], isLoading: imagesLoading } = useOptionImages(missionId, option.id);

// In JSX, above the card title/badge section:
<ImageStrip images={images} loading={imagesLoading} />
```

- [ ] **Step 4: Add i18n keys**

`frontend/src/i18n/en.json` — add under `"options"."images"`:
```json
"images": {
  "loading": "Loading images…",
  "lightboxClose": "Close",
  "gallery": "Product images"
}
```

`frontend/src/i18n/fr.json` — add under `"options"."images"`:
```json
"images": {
  "loading": "Chargement des images…",
  "lightboxClose": "Fermer",
  "gallery": "Images du produit"
}
```

`frontend/src/i18n/de.json` — add under `"options"."images"`:
```json
"images": {
  "loading": "Bilder werden geladen…",
  "lightboxClose": "Schließen",
  "gallery": "Produktbilder"
}
```

- [ ] **Step 5: Build + type-check**

```bash
cd frontend && npm run build 2>&1 | tail -20
cd frontend && npx tsc --noEmit 2>&1 | head -20
# Expected: no errors
```

- [ ] **Step 6: Commit**

```bash
git add frontend/src/components/options/OptionCard.tsx \
        frontend/src/pages/OptionsExplorer.tsx \
        frontend/src/i18n/en.json \
        frontend/src/i18n/fr.json \
        frontend/src/i18n/de.json
git commit -m "feat(images): wire ImageStrip into OptionCard + i18n keys (EN/FR/DE)"
```

---

## Chunk 5: Final verification

- [ ] **Full backend test run**

```bash
cd backend && go test ./... -count=1 2>&1 | grep -E "FAIL|ok"
# Expected: all "ok" lines, no "FAIL"
```

- [ ] **Full frontend build**

```bash
cd frontend && npm run build 2>&1 | tail -10
cd frontend && npx tsc --noEmit 2>&1 | head -20
# Expected: no errors
```

- [ ] **go vet**

```bash
cd backend && go vet ./...
# Expected: no output (no warnings)
```

- [ ] **Verify de.json has image keys**

```bash
grep -A4 '"images"' frontend/src/i18n/de.json
# Expected: loading, lightboxClose, gallery keys present
```

- [ ] **Final commit**

```bash
git add -u
git commit -m "feat(images): product images — MinIO + imagefetch package + frontend filmstrip (EN/FR/DE)"
```
