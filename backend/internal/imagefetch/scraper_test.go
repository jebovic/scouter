package imagefetch_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/jibei/scouter/internal/imagefetch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScraper_ExtractsOGImage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/img.jpg" {
			w.Header().Set("Content-Type", "image/jpeg")
			w.Write(make([]byte, 1024))
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

func TestScraper_SkipsExistingURLs(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/img.jpg" {
			w.Header().Set("Content-Type", "image/jpeg")
			w.Write(make([]byte, 1024))
			return
		}
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprintf(w, `<html><head>
            <meta property="og:image" content="%s/img.jpg">
        </head><body></body></html>`, "http://"+r.Host)
	}))
	defer srv.Close()

	existing := map[string]bool{srv.URL + "/img.jpg": true}
	s := imagefetch.NewScraper()
	imgs, err := s.Fetch(t.Context(), srv.URL, existing)
	require.NoError(t, err)
	assert.Empty(t, imgs, "image already in DB should be skipped via existingURLs")
}
