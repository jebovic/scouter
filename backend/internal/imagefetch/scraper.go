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

// Scraper fetches HTML pages and extracts product image candidates.
type Scraper struct{ client *http.Client }

// NewScraper returns a Scraper with a default HTTP client.
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
// og:image, twitter:image, JSON-LD image, <img width>=300>
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
		if m[1] != "" {
			add(m[1])
		} else if m[2] != "" {
			add(m[2])
		}
	}
	// twitter:image
	for _, m := range twitterImageRE.FindAllStringSubmatch(html, -1) {
		if m[1] != "" {
			add(m[1])
		} else if m[2] != "" {
			add(m[2])
		}
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
