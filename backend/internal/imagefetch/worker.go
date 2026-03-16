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
			if delErr := w.uploader.Delete(ctx, key); delErr != nil {
				log.Printf("imagefetch: cleanup orphaned object %s: %v", key, delErr)
			}
		}
	}
}
