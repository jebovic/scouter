package embedding_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/embedding"
	"github.com/jibei/scouter/internal/option"
)

// ── fakes ────────────────────────────────────────────────────────────────────

type fakeEmbedder struct {
	vec []float64
	err error
}

func (f *fakeEmbedder) Embed(_ context.Context, _ string) ([]float64, error) {
	return f.vec, f.err
}
func (f *fakeEmbedder) EmbedBatch(_ context.Context, texts []string) ([][]float64, error) {
	out := make([][]float64, len(texts))
	for i := range texts {
		out[i] = f.vec
	}
	return out, f.err
}

type fakeOptRepo struct {
	opt *option.Option
	err error
}

func (f *fakeOptRepo) GetByID(_ context.Context, _ uuid.UUID) (*option.Option, error) {
	return f.opt, f.err
}
func (f *fakeOptRepo) ListByMission(_ context.Context, _ uuid.UUID) ([]option.Option, error) {
	return nil, nil
}
func (f *fakeOptRepo) ListByMissionPaged(_ context.Context, _ uuid.UUID, _ *time.Time, _ int) ([]option.Option, error) {
	return nil, nil
}
func (f *fakeOptRepo) Create(_ context.Context, o option.Option) (*option.Option, error) {
	return &o, nil
}
func (f *fakeOptRepo) Update(_ context.Context, _ uuid.UUID, _ option.UpdateRequest) (*option.Option, error) {
	return nil, nil
}
func (f *fakeOptRepo) Delete(_ context.Context, _ uuid.UUID) error        { return nil }
func (f *fakeOptRepo) DeleteByMission(_ context.Context, _ uuid.UUID) error { return nil }
func (f *fakeOptRepo) Pin(_ context.Context, _ uuid.UUID) (*option.Option, error) {
	return nil, nil
}
func (f *fakeOptRepo) Reject(_ context.Context, _ uuid.UUID, _ option.RejectRequest) (*option.Option, error) {
	return nil, nil
}
func (f *fakeOptRepo) Unreject(_ context.Context, _ uuid.UUID) (*option.Option, error) {
	return nil, nil
}
func (f *fakeOptRepo) DeletePinned(_ context.Context, _ uuid.UUID) error { return nil }

type fakeEmbedRepo struct {
	ids    []uuid.UUID
	stored map[uuid.UUID][]float64
}

func newFakeEmbedRepo() *fakeEmbedRepo {
	return &fakeEmbedRepo{stored: make(map[uuid.UUID][]float64)}
}

func (f *fakeEmbedRepo) SetEmbedding(_ context.Context, id uuid.UUID, vec []float64) error {
	f.stored[id] = vec
	return nil
}
func (f *fakeEmbedRepo) FindUnembeddedIDs(_ context.Context, _ int) ([]uuid.UUID, error) {
	return f.ids, nil
}

// ── tests ────────────────────────────────────────────────────────────────────

func TestWorker_Submit_Process(t *testing.T) {
	id := uuid.New()
	opt := &option.Option{ID: id, Name: "Test", Category: "Cat", Notes: "notes"}
	vec := []float64{0.1, 0.2, 0.3}

	embedder := &fakeEmbedder{vec: vec}
	optRepo := &fakeOptRepo{opt: opt}
	embedRepo := newFakeEmbedRepo()

	w := embedding.NewWorker(embedder, optRepo, embedRepo)
	ctx, cancel := context.WithCancel(context.Background())
	w.Start(ctx)

	if !w.Submit(id) {
		t.Fatal("Submit returned false unexpectedly")
	}

	// Give worker goroutines time to process.
	time.Sleep(50 * time.Millisecond)
	cancel()
	w.Wait()

	if got, ok := embedRepo.stored[id]; !ok {
		t.Fatal("embedding was not stored")
	} else if len(got) != len(vec) {
		t.Errorf("stored embedding len = %d, want %d", len(got), len(vec))
	}
}

func TestWorker_Backfill(t *testing.T) {
	id1, id2 := uuid.New(), uuid.New()
	opt := &option.Option{ID: id1, Name: "A", Category: "B"}

	embedder := &fakeEmbedder{vec: []float64{1, 2}}
	optRepo := &fakeOptRepo{opt: opt}
	embedRepo := newFakeEmbedRepo()
	embedRepo.ids = []uuid.UUID{id1, id2}

	w := embedding.NewWorker(embedder, optRepo, embedRepo)
	n := w.Backfill(context.Background(), 10)
	if n != 2 {
		t.Errorf("Backfill returned %d, want 2", n)
	}
}

func TestWorker_Submit_ChannelFull(t *testing.T) {
	embedder := &fakeEmbedder{vec: []float64{1}}
	optRepo := &fakeOptRepo{opt: &option.Option{Name: "X", Category: "Y"}}
	embedRepo := newFakeEmbedRepo()

	w := embedding.NewWorker(embedder, optRepo, embedRepo)
	// Don't call Start so goroutines don't drain the channel.
	// Fill the channel (capacity 100).
	for i := 0; i < 100; i++ {
		w.Submit(uuid.New())
	}
	// 101st submit should return false (channel full).
	if w.Submit(uuid.New()) {
		t.Error("Submit should return false when channel is full")
	}
}
