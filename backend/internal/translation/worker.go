package translation

import (
	"context"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/option"
)

const (
	workerChannelCap = 256
	workerCount      = 2
)

// Worker processes translation jobs in the background.
type Worker struct {
	jobs       chan uuid.UUID
	translator *Translator
	repo       option.Repository
	locales    []string
	wg         sync.WaitGroup
	// testHook is called instead of the real translate+save pipeline when set (tests only)
	testHook func(uuid.UUID)
}

// NewWorker creates a Worker.
// Pass testHook=nil in production; tests may provide a hook for fast verification.
func NewWorker(translator *Translator, repo option.Repository, locales []string, testHook func(uuid.UUID)) *Worker {
	return &Worker{
		jobs:       make(chan uuid.UUID, workerChannelCap),
		translator: translator,
		repo:       repo,
		locales:    locales,
		testHook:   testHook,
	}
}

// Submit enqueues an option ID for translation.
// Returns false and logs a warning when the channel is full (drop-on-full policy).
func (w *Worker) Submit(id uuid.UUID) bool {
	select {
	case w.jobs <- id:
		return true
	default:
		log.Printf("translation: worker channel full, dropping %s", id)
		return false
	}
}

// Jobs returns the send-only end of the job channel so callers can submit
// without holding a reference to the full Worker struct.
func (w *Worker) Jobs() chan<- uuid.UUID { return w.jobs }

// Start launches workerCount goroutines that drain the jobs channel.
func (w *Worker) Start(ctx context.Context) {
	for range workerCount {
		w.wg.Add(1)
		go func() {
			defer w.wg.Done()
			for {
				select {
				case id, ok := <-w.jobs:
					if !ok {
						return
					}
					if w.testHook != nil {
						w.testHook(id)
						continue
					}
					w.process(ctx, id)
				case <-ctx.Done():
					return
				}
			}
		}()
	}
}

// Wait blocks until all goroutines have exited.
func (w *Worker) Wait() {
	w.wg.Wait()
}

func (w *Worker) process(ctx context.Context, id uuid.UUID) {
	o, err := w.repo.GetByID(ctx, id)
	if err != nil || o == nil {
		log.Printf("translation: cannot load option %s: %v", id, err)
		return
	}
	for _, locale := range w.locales {
		raw, err := w.translator.Translate(ctx, *o, locale)
		if err != nil {
			log.Printf("translation: translate %s→%s failed: %v", id, locale, err)
			continue
		}
		if err := w.repo.SaveTranslations(ctx, id, locale, raw); err != nil {
			log.Printf("translation: save %s→%s failed: %v", id, locale, err)
		}
	}
}
