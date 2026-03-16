# Product Images & Multilingual Generated Content

**Date:** 2026-03-16
**Status:** Approved
**Features:** Product image gallery (MinIO), i18n of LLM-generated option content

---

## Context

Two gaps exist in the current option indexing pipeline:

1. **No images** — options store only text (name, attributes, notes, URL). Product images would make the OptionsExplorer significantly more useful for visual comparison.
2. **English-only generated content** — research and pricing agents generate all content in English, even when the user's locale is French or German. The app UI is fully i18n'd, but the generated data is not.

---

## Feature 1: Product Images

### Goal

Download and store product images alongside each indexed option. Images are scraped from the option's own product URL, stored in MinIO, and displayed as a horizontal filmstrip in the OptionCard.

### Data Model

New migration **023**: `option_images` table.

```sql
-- 023_option_images.up.sql
CREATE TABLE option_images (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    option_id    UUID NOT NULL REFERENCES options(id) ON DELETE CASCADE,
    minio_key    TEXT NOT NULL,          -- "options/{option_id}/{uuid}.jpg"
    content_type TEXT NOT NULL,          -- "image/jpeg", "image/webp", etc.
    width        INT,
    height       INT,
    source_url   TEXT,                   -- original URL (for deduplication)
    sort_order   INT NOT NULL DEFAULT 0, -- 0 = primary/cover image
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX ON option_images(option_id);

-- 023_option_images.down.sql
DROP TABLE IF EXISTS option_images;
```

`ON DELETE CASCADE` ensures DB records are removed when an option is deleted. MinIO object cleanup is handled separately by the purge job.

### Infrastructure: MinIO

New service in `deployment/docker-compose.yml`:

```yaml
minio:
  image: minio/minio
  command: server /data --console-address ":9001"
  environment:
    MINIO_ROOT_USER: ${MINIO_ACCESS_KEY}
    MINIO_ROOT_PASSWORD: ${MINIO_SECRET_KEY}
  volumes:
    - minio_data:/data
  labels:
    - "traefik.enable=true"
    - "traefik.http.routers.minio.rule=Host(`minio.dev.local`)"
    - "traefik.http.routers.minio.entrypoints=websecure"
    - "traefik.http.routers.minio.tls=true"
    - "traefik.http.services.minio.loadbalancer.server.port=9000"
```

The MinIO Go client is configured with `Secure: false` (plain HTTP on the internal Docker network). The backend generates presigned URLs using `MINIO_PUBLIC_URL` (e.g. `https://minio.dev.local`) so the browser receives HTTPS URLs routed via Traefik, while the backend communicates with MinIO over plain HTTP internally.

New env vars (added to `.env.example`):

| Variable | Default | Notes |
|---|---|---|
| `MINIO_ENDPOINT` | `minio:9000` | internal Docker hostname (plain HTTP) |
| `MINIO_PUBLIC_URL` | `https://minio.dev.local` | used when generating presigned URLs |
| `MINIO_ACCESS_KEY` | — | required |
| `MINIO_SECRET_KEY` | — | required |
| `MINIO_BUCKET` | `scouter-images` | auto-created on startup |
| `MINIO_QUOTA_MB` | `2048` | 2 GiB storage quota |

### Backend Package: `internal/imagefetch/`

**`scraper.go`**
- Fetches `option.url` with a browser User-Agent + 10s timeout
- Extracts image candidates in priority order: `og:image`, `twitter:image`, JSON-LD `image`, `<img>` tags with `width ≥ 300`
- Deduplicates against existing `option_images.source_url` for the option
- Downloads each candidate (max 5 per option, max 5 MB each), reads content type
- Returns `[]FetchedImage{Bytes, ContentType, Width, Height, SourceURL}`
- Note: no `robots.txt` compliance or rate-limiting (personal tool, out of scope)

**`uploader.go`**
- MinIO client: `endpoint = MINIO_ENDPOINT`, `Secure: false` (TLS terminated by Traefik)
- Key format: `options/{option_id}/{uuid}.{ext}`
- Creates bucket on first use if absent
- After successful upload: inserts `option_images` row
- Presigned URL host overridden using `MINIO_PUBLIC_URL` before returning to callers

**`worker.go`**
- Buffered channel `chan uuid.UUID` (capacity 256; higher than the embedding worker's 100 because image scraping is significantly slower and the queue can grow during bulk research)
- 2 goroutines process the queue concurrently
- Research agent sends each saved option ID to this channel after DB persist
- Started at server boot alongside the scheduler

**`purge.go`** (runs nightly via `robfig/cron`)
- Iterates all MinIO objects via `ListObjectsV2`, summing their `Size` fields client-side (there is no atomic bucket-size API in the S3/MinIO protocol)
- If total exceeds `MINIO_QUOTA_MB`: deletes oldest objects (sorted by `LastModified` ascending) and their `option_images` rows until usage drops below 90% of quota

### API Endpoints

All image endpoints are nested under the existing option route hierarchy:

| Method | Path | Description |
|---|---|---|
| `GET` | `/api/missions/:missionId/options/:optionId/images` | List images |
| `PATCH` | `/api/missions/:missionId/options/:optionId/images/:imageId` | Update `sort_order` |
| `DELETE` | `/api/missions/:missionId/options/:optionId/images/:imageId` | Delete image |

`GET` response shape:
```json
[
  {
    "id": "...",
    "url": "https://minio.dev.local/...",
    "contentType": "image/jpeg",
    "width": 800,
    "height": 600,
    "sortOrder": 0
  }
]
```

Presigned URLs use `MINIO_PUBLIC_URL` as host, TTL 1h. The frontend `useOptionImages` hook uses `staleTime: 50min` with `refetchOnWindowFocus: true` to guarantee URLs are refreshed before the 1h expiry.

### Frontend

**`useOptionImages(missionId, optionId)`** hook
- `GET /api/missions/:missionId/options/:optionId/images` via TanStack Query
- `staleTime: 50min`, `refetchOnWindowFocus: true` (ensures presigned URLs never expire in active sessions)
- Polls every 30s (max 10 retries) while `images.length === 0` to catch the async worker finishing

**OptionCard image strip**
- Renders above card content when `images.length > 0`
- Horizontal filmstrip: thumbnails 80×60px, `object-fit: cover`, `border-radius: 8px`
- Max 4 thumbnails visible; overflow shows "+N" chip
- Clicking any thumbnail opens a full-size lightbox
- While loading (poll active): 3-tile animated skeleton strip (reuses existing `Skeleton` component)
- If scraping found no images: strip absent (no empty state)

**New i18n keys** (add to `en.json`, `fr.json`, **and `de.json`**):
```
options.images.loading       → "Loading images…" / "Chargement des images…" / "Bilder werden geladen…"
options.images.lightboxClose → "Close" / "Fermer" / "Schließen"
```

The image strip component must use a co-located `.module.css` file per project CSS conventions (not inline `style={{}}`), even though the surrounding `OptionCard` predates CSS Modules.

---

## Feature 2: Multilingual Generated Content

### Goal

Research and pricing agents generate all content in English. A background translation worker produces translated versions of all text fields for each configured locale (default: `fr`, `de`), stored in a `translations` JSONB column on the `options` table. The frontend renders translated content when available, falling back to English with a "translating…" indicator.

### Data Model

New migration **024**: `translations` column on `options`.

```sql
-- 024_option_translations.up.sql
ALTER TABLE options ADD COLUMN translations JSONB;

-- 024_option_translations.down.sql
ALTER TABLE options DROP COLUMN translations;
```

Shape of `translations`:
```json
{
  "fr": {
    "name": "...",
    "notes": "...",
    "warnings": ["...", "..."],
    "category": "...",
    "attributes": [
      { "label": "...", "value": "..." }
    ]
  },
  "de": { ... }
}
```

**Translated attribute rules:**
- `type == "text"`: both `label` and `value` are translated
- `type == "price"` / `"score"` / `"boolean"`: `value` is copied verbatim; only `label` is translated
- Attribute array order is preserved

**Save strategy:** `UPDATE options SET translations = translations || $1 WHERE id = $2`

The PostgreSQL `||` operator performs a shallow (top-level key) merge on JSONB. Since the translation blob is keyed by locale (`{"fr": {...}}`), this correctly replaces the full French translation atomically while preserving other locale keys already present. This is intentional — we want locale-level replacement, not field-level merging.

### Backend Package: `internal/translation/`

**`translator.go`**
- Accepts `Option` + target locale string (`"fr"`, `"de"`) → returns translated JSONB blob
- Builds compact prompt: sends `name`, `notes`, `warnings[]`, `category`, and qualifying `{label, value}` attribute pairs as JSON
- Requests translated JSON in the same schema
- Uses `SmartRouter` with `WithRequestOpts(fast)` — no tool use, just JSON completion
- Uses `RetryAsJSON` fallback (already in codebase)

**`fields.go`**
- `TranslatableFields(option Option) TranslationInput` — extracts only translatable content
- `MergeTranslation(option Option, locale string, blob json.RawMessage) Option` — merges translated blob immutably
- Both functions are pure and easily unit-tested without LLM dependency

**`worker.go`**
- Buffered channel `chan uuid.UUID` (capacity 256), 2 goroutines
- Research agent sends option IDs here after DB persist (same pattern as image worker)
- Translates into all locales from `SUPPORTED_LOCALES` env var (default: `"fr,de"`)
- Saves result: `UPDATE options SET translations = translations || $1 WHERE id = $2`

### New Env Var

| Variable | Default | Notes |
|---|---|---|
| `SUPPORTED_LOCALES` | `fr,de` | Comma-separated locales to translate into |

### API Changes

- `Option` response DTO gains optional `translations` field (passed through from DB column)
- New endpoint: `POST /api/missions/:missionId/options/:optionId/retranslate`
  - Returns `202 Accepted` with empty body — re-queues the option for translation
  - Frontend calls `queryClient.invalidateQueries(['options', missionId])` after receiving 202 to trigger a refetch once translations arrive (via polling on the options query)

### Frontend

**`useTranslatedOption(option: Option): Option`** hook
- Reads `i18n.language` (e.g. `"fr"`)
- If `option.translations?.[lang]` exists: immutably merges translated fields over English fields
- Returns a merged `Option` (same type — the translation shape is structurally identical to the option's text fields; no new Zod schema needed)
- Used in `OptionCard`, `AttributeRenderer`, `ComparisonTable`

**Translation pending indicator**
- If `i18n.language !== "en"` and `option.translations?.[lang]` is absent: renders a `🌐` badge with tooltip "Translation in progress" on the OptionCard
- Badge disappears once the options query refetch detects `translations[lang]` has been populated by the worker

**New i18n keys** (add to `en.json`, `fr.json`, **and `de.json`**):
```
options.translating → "Translating…" / "Traduction en cours…" / "Wird übersetzt…"
```

Note: the existing `OptionSchema` in `src/api/options.ts` and the `Option` type in `src/types/option.ts` must gain an optional `translations` field (`translations?: Record<string, TranslationBlob>`) to pass through the new DB column. No separate schema is needed.

---

## Implementation Order

These two features are independent and can be phased separately:

**Phase A — Product Images**
1. Migration 023: `option_images` table (up + down)
2. Docker Compose: MinIO service + volume + Traefik labels
3. `internal/imagefetch/` package (scraper, uploader, worker, purge)
4. Wire worker into server boot + research agent
5. API endpoints (nested under mission/option routes)
6. Frontend: `useOptionImages`, image strip in `OptionCard`, lightbox, i18n keys

**Phase B — Multilingual Generated Content**
1. Migration 024: `translations` column (up + down)
2. `internal/translation/` package (translator, fields, worker)
3. Wire worker into server boot + research agent
4. `POST /api/missions/:missionId/options/:optionId/retranslate` (202 Accepted)
5. Frontend: `useTranslatedOption` hook, translation pending badge, query invalidation
6. Add i18n keys to `en.json` + `fr.json`

---

## Out of Scope

- Translating pricing agent price listings (merchant names, descriptions)
- Image editing or resizing beyond what the scraper naturally provides
- CDN in front of MinIO
- User-uploaded images
- `robots.txt` compliance or rate-limiting in the image scraper
