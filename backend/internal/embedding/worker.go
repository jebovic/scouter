package embedding

import (
	"context"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/llm"
	"github.com/jibei/scouter/internal/option"
)

const (
	workerCount   = 2
	channelCap    = 100
	backfillLimit = 200
)

// Worker is an async embedding processor. It receives option IDs over a buffered
// channel and embeds them using the configured EmbedProvider without blocking callers.
type Worker struct {
	jobs     chan uuid.UUID
	embedder llm.EmbedProvider
	optRepo  option.Repository
	embedRepo Repository
	wg       sync.WaitGroup
}

// NewWorker constructs an embedding Worker. Call Start to begin processing.
func NewWorker(embedder llm.EmbedProvider, optRepo option.Repository, embedRepo Repository) *Worker {
	return &Worker{
		jobs:      make(chan uuid.UUID, channelCap),
		embedder:  embedder,
		optRepo:   optRepo,
		embedRepo: embedRepo,
	}
}

// Jobs returns a send-only view of the job channel for external producers.
func (w *Worker) Jobs() chan<- uuid.UUID {
	return w.jobs
}

// Submit enqueues an option ID for embedding. Returns false if the channel is full.
func (w *Worker) Submit(id uuid.UUID) bool {
	select {
	case w.jobs <- id:
		return true
	default:
		log.Printf("embedding: worker channel full, dropping %s", id)
		return false
	}
}

// Start launches workerCount background goroutines that process jobs until ctx is
// cancelled. Call Wait (or use the returned WaitGroup) to wait for clean shutdown.
func (w *Worker) Start(ctx context.Context) {
	for i := 0; i < workerCount; i++ {
		w.wg.Add(1)
		go func() {
			defer w.wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case id, ok := <-w.jobs:
					if !ok {
						return
					}
					w.process(ctx, id)
				}
			}
		}()
	}
}

// Wait blocks until all worker goroutines have exited.
func (w *Worker) Wait() {
	w.wg.Wait()
}

// Backfill enqueues up to limit un-embedded options for (re)embedding.
// It is called at startup to catch any options that were persisted without embeddings.
func (w *Worker) Backfill(ctx context.Context, limit int) int {
	ids, err := w.embedRepo.FindUnembeddedIDs(ctx, limit)
	if err != nil {
		log.Printf("embedding: backfill find: %v", err)
		return 0
	}
	n := 0
	for _, id := range ids {
		if w.Submit(id) {
			n++
		}
	}
	return n
}

// process fetches the option, builds embed text, calls the embedder, and persists.
func (w *Worker) process(ctx context.Context, id uuid.UUID) {
	opt, err := w.optRepo.GetByID(ctx, id)
	if err != nil {
		log.Printf("embedding: get option %s: %v", id, err)
		return
	}

	if opt == nil {
		return // option deleted between submit and process
	}

	text := BuildEmbedText(*opt)
	if text == "" {
		log.Printf("embedding: empty text for option %s, skipping", id)
		return
	}

	vec, err := w.embedder.Embed(ctx, text)
	if err != nil {
		log.Printf("embedding: embed option %s: %v", id, err)
		return
	}

	if err := w.embedRepo.SetEmbedding(ctx, id, vec); err != nil {
		log.Printf("embedding: set embedding for %s: %v", id, err)
	}
}
