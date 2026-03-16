package translation_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/translation"
	"github.com/stretchr/testify/assert"
)

func TestWorker_SubmitAndDrain(t *testing.T) {
	done := make(chan uuid.UUID, 1)
	w := translation.NewWorker(nil, nil, []string{"fr"}, func(id uuid.UUID) {
		done <- id
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	w.Start(ctx)

	id := uuid.New()
	ok := w.Submit(id)
	assert.True(t, ok)

	select {
	case got := <-done:
		assert.Equal(t, id, got)
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for translation job to be processed")
	}

	cancel()
	w.Wait()
}

func TestWorker_DropWhenFull(t *testing.T) {
	// Slow handler that blocks until we release
	release := make(chan struct{})
	w := translation.NewWorker(nil, nil, []string{"fr"}, func(id uuid.UUID) {
		<-release
	})
	ctx, cancel := context.WithCancel(context.Background())
	w.Start(ctx)

	// Fill channel (cap 256) + 2 goroutines = submit 259 to guarantee a drop
	dropped := false
	for i := 0; i < 260; i++ {
		if !w.Submit(uuid.New()) {
			dropped = true
			break
		}
	}
	assert.True(t, dropped)

	close(release)
	cancel()
	w.Wait()
}
