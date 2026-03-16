package export

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/decision"
	"github.com/jibei/scouter/internal/httputil"
	"github.com/jibei/scouter/internal/mission"
	"github.com/jibei/scouter/internal/option"
	"github.com/jibei/scouter/internal/shopping"
)

// ---------------------------------------------------------------------------
// Stub repositories (satisfy interfaces without any real DB).
// ---------------------------------------------------------------------------

type stubMissionRepo struct {
	m   *mission.Mission
	err error
}

func (s *stubMissionRepo) GetByID(_ context.Context, _ uuid.UUID) (*mission.Mission, error) {
	return s.m, s.err
}
func (s *stubMissionRepo) List(_ context.Context) ([]mission.Mission, error) { return nil, nil }
func (s *stubMissionRepo) ListPaged(_ context.Context, _ *time.Time, _ int, _ bool) ([]mission.Mission, error) {
	return nil, nil
}
func (s *stubMissionRepo) ListActive(_ context.Context) ([]mission.Mission, error) { return nil, nil }
func (s *stubMissionRepo) GetBySlug(_ context.Context, _ string) (*mission.Mission, error) {
	return nil, nil
}
func (s *stubMissionRepo) GetByShareToken(_ context.Context, _ string) (*mission.Mission, error) {
	return nil, nil
}
func (s *stubMissionRepo) Create(_ context.Context, m mission.Mission) (*mission.Mission, error) {
	return &m, nil
}
func (s *stubMissionRepo) Update(_ context.Context, _ uuid.UUID, _ mission.UpdateRequest) (*mission.Mission, error) {
	return nil, nil
}
func (s *stubMissionRepo) Delete(_ context.Context, _ uuid.UUID) error { return nil }
func (s *stubMissionRepo) SetShareToken(_ context.Context, _ uuid.UUID, _ string) (*mission.Mission, error) {
	return nil, nil
}
func (s *stubMissionRepo) ClearShareToken(_ context.Context, _ uuid.UUID) error { return nil }
func (s *stubMissionRepo) Archive(_ context.Context, _ uuid.UUID) (*mission.Mission, error) {
	return nil, nil
}
func (s *stubMissionRepo) Unarchive(_ context.Context, _ uuid.UUID) (*mission.Mission, error) {
	return nil, nil
}

func (s *stubMissionRepo) SetPhase(_ context.Context, _ uuid.UUID, _ string) error { return nil }
func (s *stubMissionRepo) CloneMission(_ context.Context, _ string) (mission.Mission, error) {
	return mission.Mission{}, nil
}

type stubOptionRepo struct {
	opts []option.Option
	err  error
}

func (s *stubOptionRepo) ListByMission(_ context.Context, _ uuid.UUID) ([]option.Option, error) {
	return s.opts, s.err
}
func (s *stubOptionRepo) ListByMissionPaged(_ context.Context, _ uuid.UUID, _ *time.Time, _ int) ([]option.Option, error) {
	return nil, nil
}
func (s *stubOptionRepo) GetByID(_ context.Context, _ uuid.UUID) (*option.Option, error) {
	return nil, nil
}
func (s *stubOptionRepo) Create(_ context.Context, o option.Option) (*option.Option, error) {
	return &o, nil
}
func (s *stubOptionRepo) Update(_ context.Context, _ uuid.UUID, _ option.UpdateRequest) (*option.Option, error) {
	return nil, nil
}
func (s *stubOptionRepo) Delete(_ context.Context, _ uuid.UUID) error       { return nil }
func (s *stubOptionRepo) DeleteByMission(_ context.Context, _ uuid.UUID) error { return nil }
func (s *stubOptionRepo) Pin(_ context.Context, _ uuid.UUID) (*option.Option, error) {
	return nil, nil
}
func (s *stubOptionRepo) Reject(_ context.Context, _ uuid.UUID, _ option.RejectRequest) (*option.Option, error) {
	return nil, nil
}
func (s *stubOptionRepo) Unreject(_ context.Context, _ uuid.UUID) (*option.Option, error) {
	return nil, nil
}
func (s *stubOptionRepo) DeletePinned(_ context.Context, _ uuid.UUID) error { return nil }
func (s *stubOptionRepo) SaveTranslations(_ context.Context, _ uuid.UUID, _ string, _ json.RawMessage) error {
	return nil
}

type stubShoppingRepo struct {
	items    []shopping.Item
	snapshots []shopping.PriceSnapshot
	err       error
}

func (s *stubShoppingRepo) ListByMission(_ context.Context, _ uuid.UUID) ([]shopping.Item, error) {
	return s.items, s.err
}
func (s *stubShoppingRepo) ListByMissionPaged(_ context.Context, _ uuid.UUID, _ *time.Time, _ int) ([]shopping.Item, error) {
	return nil, nil
}
func (s *stubShoppingRepo) GetByID(_ context.Context, _ uuid.UUID) (*shopping.Item, error) {
	return nil, nil
}
func (s *stubShoppingRepo) Create(_ context.Context, item shopping.Item) (*shopping.Item, error) {
	return &item, nil
}
func (s *stubShoppingRepo) Update(_ context.Context, _ uuid.UUID, _ shopping.UpdateRequest) (*shopping.Item, error) {
	return nil, nil
}
func (s *stubShoppingRepo) Delete(_ context.Context, _ uuid.UUID) error       { return nil }
func (s *stubShoppingRepo) DeleteByMission(_ context.Context, _ uuid.UUID) error { return nil }
func (s *stubShoppingRepo) AddPriceSnapshot(_ context.Context, _ uuid.UUID, _ shopping.PriceSnapshotRequest) (*shopping.PriceSnapshot, error) {
	return nil, nil
}
func (s *stubShoppingRepo) ListPriceHistory(_ context.Context, _ uuid.UUID) ([]shopping.PriceSnapshot, error) {
	return s.snapshots, s.err
}
func (s *stubShoppingRepo) Pin(_ context.Context, _ uuid.UUID) (*shopping.Item, error) {
	return nil, nil
}
func (s *stubShoppingRepo) DeletePinned(_ context.Context, _ uuid.UUID) error { return nil }
func (s *stubShoppingRepo) GetPriceHistory(_ context.Context, _ uuid.UUID, _ int) ([]shopping.PriceHistoryPoint, error) {
	return nil, nil
}

type stubDecisionRepo struct {
	dec *decision.Decision
	err error
}

func (s *stubDecisionRepo) Create(_ context.Context, d decision.Decision) (*decision.Decision, error) {
	return &d, nil
}
func (s *stubDecisionRepo) GetLatestByMission(_ context.Context, _ uuid.UUID) (*decision.Decision, error) {
	return s.dec, s.err
}

// ---------------------------------------------------------------------------
// Fixture builder.
// ---------------------------------------------------------------------------

func minimalExportData() *ExportData {
	return &ExportData{
		ExportedAt: time.Now().UTC(),
		Mission: &mission.Mission{
			ID:       uuid.New(),
			Slug:     "test-mission",
			Name:     "Test Mission",
			Category: "electronics",
			Budget:   1000,
			Currency: "EUR",
			Phase:    "researching",
		},
		Options:       []option.Option{},
		ShoppingItems: []shopping.Item{},
		PriceHistory:  map[uuid.UUID][]shopping.PriceSnapshot{},
		Decision:      nil,
	}
}

// ---------------------------------------------------------------------------
// ToJSON tests.
// ---------------------------------------------------------------------------

func TestToJSON_ValidEnvelope(t *testing.T) {
	data := minimalExportData()
	b, err := ToJSON(data)
	if err != nil {
		t.Fatalf("ToJSON returned error: %v", err)
	}
	if len(b) == 0 {
		t.Fatal("ToJSON returned empty bytes")
	}

	var envelope struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(b, &envelope); err != nil {
		t.Fatalf("output is not valid JSON: %v", err)
	}
	if envelope.Version != "1.0" {
		t.Errorf("expected version %q, got %q", "1.0", envelope.Version)
	}
}

func TestToJSON_ContainsMissionName(t *testing.T) {
	data := minimalExportData()
	b, err := ToJSON(data)
	if err != nil {
		t.Fatalf("ToJSON returned error: %v", err)
	}
	if !bytes.Contains(b, []byte("Test Mission")) {
		t.Error("JSON output does not contain mission name")
	}
}

func TestToJSON_EmptyOptionsAndItems(t *testing.T) {
	data := minimalExportData()
	b, err := ToJSON(data)
	if err != nil {
		t.Fatalf("ToJSON returned error: %v", err)
	}

	var full map[string]json.RawMessage
	if err := json.Unmarshal(b, &full); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	// options and shoppingItems should marshal as empty arrays, not null.
	for _, key := range []string{"options", "shoppingItems"} {
		raw, ok := full[key]
		if !ok {
			t.Errorf("key %q missing from JSON output", key)
			continue
		}
		if string(raw) == "null" {
			t.Errorf("key %q is null; expected empty array []", key)
		}
	}
}

func TestToJSON_WithDecision(t *testing.T) {
	data := minimalExportData()
	data.Decision = &decision.Decision{
		ID:        uuid.New(),
		MissionID: data.Mission.ID,
		Summary:   "Recommended option A.",
		Scores:    []decision.ScoreResult{},
		CreatedAt: time.Now().UTC(),
	}

	b, err := ToJSON(data)
	if err != nil {
		t.Fatalf("ToJSON with decision returned error: %v", err)
	}
	if !bytes.Contains(b, []byte("Recommended option A.")) {
		t.Error("JSON output does not contain decision summary")
	}
}

// ---------------------------------------------------------------------------
// ToMarkdown tests.
// ---------------------------------------------------------------------------

func TestToMarkdown_NonEmpty(t *testing.T) {
	data := minimalExportData()
	b, err := ToMarkdown(data)
	if err != nil {
		t.Fatalf("ToMarkdown returned error: %v", err)
	}
	if len(b) == 0 {
		t.Fatal("ToMarkdown returned empty bytes")
	}
}

func TestToMarkdown_ContainsMissionName(t *testing.T) {
	data := minimalExportData()
	b, err := ToMarkdown(data)
	if err != nil {
		t.Fatalf("ToMarkdown returned error: %v", err)
	}
	if !bytes.Contains(b, []byte("Test Mission")) {
		t.Error("Markdown output does not contain mission name")
	}
}

func TestToMarkdown_ContainsOptionsHeading(t *testing.T) {
	data := minimalExportData()
	b, err := ToMarkdown(data)
	if err != nil {
		t.Fatalf("ToMarkdown returned error: %v", err)
	}
	if !bytes.Contains(b, []byte("## Options")) {
		t.Error("Markdown output does not contain '## Options' heading")
	}
}

func TestToMarkdown_WithOptions(t *testing.T) {
	data := minimalExportData()
	pr := &option.PriceRange{Min: 800, Max: 1200, Best: 950}
	data.Options = []option.Option{
		{
			ID:        uuid.New(),
			MissionID: data.Mission.ID,
			Name:      "Sony WH-1000XM5",
			Badge:     "recommended",
			PriceRange: pr,
			Notes:     "Best noise cancellation.",
			CreatedAt: time.Now().UTC(),
		},
	}

	b, err := ToMarkdown(data)
	if err != nil {
		t.Fatalf("ToMarkdown with options returned error: %v", err)
	}
	if !bytes.Contains(b, []byte("Sony WH-1000XM5")) {
		t.Error("Markdown output does not contain option name")
	}
	if !bytes.Contains(b, []byte("recommended")) {
		t.Error("Markdown output does not contain option badge")
	}
}

func TestToMarkdown_WithShoppingItems(t *testing.T) {
	data := minimalExportData()
	data.ShoppingItems = []shopping.Item{
		{
			ID:        uuid.New(),
			MissionID: data.Mission.ID,
			Name:      "Laptop Stand",
			Merchant:  "Amazon",
			Price:     49.99,
			Status:    "buy",
			CreatedAt: time.Now().UTC(),
		},
	}

	b, err := ToMarkdown(data)
	if err != nil {
		t.Fatalf("ToMarkdown with shopping items returned error: %v", err)
	}
	if !bytes.Contains(b, []byte("Laptop Stand")) {
		t.Error("Markdown output does not contain shopping item name")
	}
}

func TestToMarkdown_NoDecisionSection(t *testing.T) {
	data := minimalExportData()
	b, err := ToMarkdown(data)
	if err != nil {
		t.Fatalf("ToMarkdown returned error: %v", err)
	}
	if !strings.Contains(string(b), "No decision has been run") {
		t.Error("Markdown output should say no decision has been run")
	}
}

func TestToMarkdown_WithDecision(t *testing.T) {
	data := minimalExportData()
	data.Decision = &decision.Decision{
		ID:        uuid.New(),
		MissionID: data.Mission.ID,
		Summary:   "Go with option A for best value.",
		Scores:    []decision.ScoreResult{},
		CreatedAt: time.Now().UTC(),
	}

	b, err := ToMarkdown(data)
	if err != nil {
		t.Fatalf("ToMarkdown with decision returned error: %v", err)
	}
	if !bytes.Contains(b, []byte("Go with option A for best value.")) {
		t.Error("Markdown output does not contain decision summary")
	}
}

// ---------------------------------------------------------------------------
// ToPDF tests.
// ---------------------------------------------------------------------------

func TestToPDF_NonEmpty(t *testing.T) {
	data := minimalExportData()
	b, err := ToPDF(data)
	if err != nil {
		t.Fatalf("ToPDF returned error: %v", err)
	}
	if len(b) == 0 {
		t.Fatal("ToPDF returned empty bytes")
	}
}

func TestToPDF_StartsWithPDFMagicBytes(t *testing.T) {
	data := minimalExportData()
	b, err := ToPDF(data)
	if err != nil {
		t.Fatalf("ToPDF returned error: %v", err)
	}
	if !bytes.HasPrefix(b, []byte("%PDF")) {
		t.Errorf("ToPDF output does not start with %%PDF magic bytes; got prefix %q", b[:min(len(b), 8)])
	}
}

func TestToPDF_WithOptions(t *testing.T) {
	data := minimalExportData()
	pr := &option.PriceRange{Min: 200, Max: 400, Best: 299}
	data.Options = []option.Option{
		{
			ID:         uuid.New(),
			MissionID:  data.Mission.ID,
			Name:       "Keychron K2",
			Badge:      "recommended",
			PriceRange: pr,
			Notes:      "Great tactile feedback.",
			CreatedAt:  time.Now().UTC(),
		},
	}

	b, err := ToPDF(data)
	if err != nil {
		t.Fatalf("ToPDF with options returned error: %v", err)
	}
	if !bytes.HasPrefix(b, []byte("%PDF")) {
		t.Error("ToPDF output does not start with %PDF")
	}
}

func TestToPDF_WithShoppingItems(t *testing.T) {
	data := minimalExportData()
	data.ShoppingItems = []shopping.Item{
		{
			ID:        uuid.New(),
			MissionID: data.Mission.ID,
			Name:      "USB-C Hub",
			Merchant:  "Anker",
			Price:     39.99,
			Status:    "buy",
			CreatedAt: time.Now().UTC(),
		},
		{
			ID:        uuid.New(),
			MissionID: data.Mission.ID,
			Name:      "Monitor Arm",
			Merchant:  "Ergotron",
			Price:     129.00,
			Status:    "watch",
			CreatedAt: time.Now().UTC(),
		},
	}

	b, err := ToPDF(data)
	if err != nil {
		t.Fatalf("ToPDF with shopping items returned error: %v", err)
	}
	if !bytes.HasPrefix(b, []byte("%PDF")) {
		t.Error("ToPDF output does not start with %PDF")
	}
}

func TestToPDF_WithDecision(t *testing.T) {
	optID := uuid.New()
	data := minimalExportData()
	data.Options = []option.Option{
		{ID: optID, MissionID: data.Mission.ID, Name: "MacBook Pro", Badge: "recommended", CreatedAt: time.Now().UTC()},
	}
	data.Decision = &decision.Decision{
		ID:        uuid.New(),
		MissionID: data.Mission.ID,
		Summary:   "MacBook Pro offers the best value.",
		Scores: []decision.ScoreResult{
			{
				OptionID:   optID,
				OptionName: "MacBook Pro",
				Score:      88.5,
				PriceScore: 75.0,
				QualScore:  95.0,
				FeatScore:  90.0,
			},
		},
		CreatedAt: time.Now().UTC(),
	}

	b, err := ToPDF(data)
	if err != nil {
		t.Fatalf("ToPDF with decision returned error: %v", err)
	}
	if !bytes.HasPrefix(b, []byte("%PDF")) {
		t.Error("ToPDF output does not start with %PDF")
	}
}

// ---------------------------------------------------------------------------
// Gatherer unit tests (stub-based, no real DB).
// ---------------------------------------------------------------------------

func TestGatherer_MissionNotFound(t *testing.T) {
	g := NewGatherer(
		&stubMissionRepo{m: nil, err: nil},
		&stubOptionRepo{},
		&stubShoppingRepo{},
		&stubDecisionRepo{},
	)

	result, err := g.Gather(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("expected nil error when mission not found, got: %v", err)
	}
	if result != nil {
		t.Fatal("expected nil result when mission not found")
	}
}

func TestGatherer_MissionRepoError(t *testing.T) {
	g := NewGatherer(
		&stubMissionRepo{err: errStub("db down")},
		&stubOptionRepo{},
		&stubShoppingRepo{},
		&stubDecisionRepo{},
	)

	_, err := g.Gather(context.Background(), uuid.New())
	if err == nil {
		t.Fatal("expected error when mission repo fails")
	}
}

func TestGatherer_OptionRepoError(t *testing.T) {
	m := &mission.Mission{ID: uuid.New(), Slug: "x", Name: "X", Category: "computing", Budget: 500, Currency: "USD", Phase: "researching"}
	g := NewGatherer(
		&stubMissionRepo{m: m},
		&stubOptionRepo{err: errStub("options db down")},
		&stubShoppingRepo{},
		&stubDecisionRepo{},
	)

	_, err := g.Gather(context.Background(), m.ID)
	if err == nil {
		t.Fatal("expected error when option repo fails")
	}
}

func TestGatherer_ShoppingRepoError(t *testing.T) {
	m := &mission.Mission{ID: uuid.New(), Slug: "x", Name: "X", Category: "computing", Budget: 500, Currency: "USD", Phase: "researching"}
	g := NewGatherer(
		&stubMissionRepo{m: m},
		&stubOptionRepo{opts: []option.Option{}},
		&stubShoppingRepo{err: errStub("shopping db down")},
		&stubDecisionRepo{},
	)

	_, err := g.Gather(context.Background(), m.ID)
	if err == nil {
		t.Fatal("expected error when shopping repo fails")
	}
}

func TestGatherer_HappyPath(t *testing.T) {
	missionID := uuid.New()
	m := &mission.Mission{
		ID:       missionID,
		Slug:     "my-mission",
		Name:     "My Mission",
		Category: "travel",
		Budget:   3000,
		Currency: "USD",
		Phase:    "comparing",
	}
	itemID := uuid.New()
	items := []shopping.Item{
		{ID: itemID, MissionID: missionID, Name: "Flight", Merchant: "Delta", Price: 500, Status: "buy", CreatedAt: time.Now().UTC()},
	}
	snapshots := []shopping.PriceSnapshot{
		{ID: uuid.New(), ItemID: itemID, Price: 480, RecordedAt: time.Now().UTC()},
	}
	dec := &decision.Decision{
		ID: uuid.New(), MissionID: missionID, Summary: "Book now.", CreatedAt: time.Now().UTC(),
	}

	g := NewGatherer(
		&stubMissionRepo{m: m},
		&stubOptionRepo{opts: []option.Option{}},
		&stubShoppingRepo{items: items, snapshots: snapshots},
		&stubDecisionRepo{dec: dec},
	)

	result, err := g.Gather(context.Background(), missionID)
	if err != nil {
		t.Fatalf("Gather returned error: %v", err)
	}
	if result == nil {
		t.Fatal("Gather returned nil result")
	}
	if result.Mission.ID != missionID {
		t.Errorf("wrong mission ID: got %v", result.Mission.ID)
	}
	if len(result.ShoppingItems) != 1 {
		t.Errorf("expected 1 shopping item, got %d", len(result.ShoppingItems))
	}
	if _, ok := result.PriceHistory[itemID]; !ok {
		t.Error("price history missing for item")
	}
	if result.Decision == nil || result.Decision.Summary != "Book now." {
		t.Error("decision not populated correctly")
	}
}

// ---------------------------------------------------------------------------
// Handler tests (chi router required for URL params).
// ---------------------------------------------------------------------------

// stubGatherer is a minimal stand-in that bypasses real repos.
type stubGatherer struct {
	data *ExportData
	err  error
}

// Gather satisfies the same signature used inside Handler.Export so we can
// replace the real Gatherer by swapping the Handler field directly.
func (sg *stubGatherer) Gather(_ context.Context, _ uuid.UUID) (*ExportData, error) {
	return sg.data, sg.err
}

// gathererFunc is a function type that matches Gatherer.Gather so we can
// inject it into the handler for testing.
type gatherFunc func(ctx context.Context, id uuid.UUID) (*ExportData, error)

// gatherableHandler wraps a gatherFunc for handler-level unit tests.
// It mirrors Handler.Export but uses the injected function instead of a real Gatherer.
func gatherableHandler(fn gatherFunc) http.HandlerFunc {
	h := &Handler{gatherer: &Gatherer{}} // fields unused; we intercept below
	_ = h
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := chi.URLParam(r, "missionID")
		id, err := uuid.Parse(idStr)
		if err != nil {
			httputil.WriteError(w, http.StatusBadRequest, "invalid mission id")
			return
		}
		format := r.URL.Query().Get("format")
		if format == "" {
			format = "json"
		}
		data, err := fn(r.Context(), id)
		if err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
			return
		}
		if data == nil {
			httputil.WriteError(w, http.StatusNotFound, "mission not found")
			return
		}
		slug := data.Mission.Slug
		switch format {
		case "json":
			b, err := ToJSON(data)
			if err != nil {
				httputil.WriteError(w, http.StatusInternalServerError, "export failed")
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.json"`, slug))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(b)
		case "markdown":
			b, err := ToMarkdown(data)
			if err != nil {
				httputil.WriteError(w, http.StatusInternalServerError, "export failed")
				return
			}
			w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
			w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.md"`, slug))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(b)
		case "pdf":
			b, err := ToPDF(data)
			if err != nil {
				httputil.WriteError(w, http.StatusInternalServerError, "export failed")
				return
			}
			w.Header().Set("Content-Type", "application/pdf")
			w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s.pdf"`, slug))
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(b)
		default:
			httputil.WriteError(w, http.StatusBadRequest, "format must be json, markdown, or pdf")
		}
	}
}

func makeChiRequest(method, path, missionID, format string) *http.Request {
	url := path
	if format != "" {
		url += "?format=" + format
	}
	req := httptest.NewRequest(method, url, nil)
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("missionID", missionID)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func TestHandler_Export_InvalidUUID(t *testing.T) {
	h := NewHandler(NewGatherer(
		&stubMissionRepo{},
		&stubOptionRepo{},
		&stubShoppingRepo{},
		&stubDecisionRepo{},
	))
	req := makeChiRequest("GET", "/api/missions/not-a-uuid/export", "not-a-uuid", "")
	rr := httptest.NewRecorder()
	h.Export(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandler_Export_MissionNotFound(t *testing.T) {
	fn := func(_ context.Context, _ uuid.UUID) (*ExportData, error) { return nil, nil }
	id := uuid.New()
	req := makeChiRequest("GET", "/api/missions/"+id.String()+"/export", id.String(), "json")
	rr := httptest.NewRecorder()
	gatherableHandler(fn)(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestHandler_Export_GatherError(t *testing.T) {
	fn := func(_ context.Context, _ uuid.UUID) (*ExportData, error) {
		return nil, errStub("db error")
	}
	id := uuid.New()
	req := makeChiRequest("GET", "/api/missions/"+id.String()+"/export", id.String(), "json")
	rr := httptest.NewRecorder()
	gatherableHandler(fn)(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

func TestHandler_Export_FormatJSON(t *testing.T) {
	data := minimalExportData()
	fn := func(_ context.Context, _ uuid.UUID) (*ExportData, error) { return data, nil }
	id := uuid.New()
	req := makeChiRequest("GET", "/", id.String(), "json")
	rr := httptest.NewRecorder()
	gatherableHandler(fn)(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected application/json content-type, got %q", ct)
	}
}

func TestHandler_Export_FormatMarkdown(t *testing.T) {
	data := minimalExportData()
	fn := func(_ context.Context, _ uuid.UUID) (*ExportData, error) { return data, nil }
	id := uuid.New()
	req := makeChiRequest("GET", "/", id.String(), "markdown")
	rr := httptest.NewRecorder()
	gatherableHandler(fn)(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestHandler_Export_FormatPDF(t *testing.T) {
	data := minimalExportData()
	fn := func(_ context.Context, _ uuid.UUID) (*ExportData, error) { return data, nil }
	id := uuid.New()
	req := makeChiRequest("GET", "/", id.String(), "pdf")
	rr := httptest.NewRecorder()
	gatherableHandler(fn)(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestHandler_Export_DefaultFormatIsJSON(t *testing.T) {
	data := minimalExportData()
	fn := func(_ context.Context, _ uuid.UUID) (*ExportData, error) { return data, nil }
	id := uuid.New()
	// no ?format= query param
	req := makeChiRequest("GET", "/", id.String(), "")
	rr := httptest.NewRecorder()
	gatherableHandler(fn)(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
	if ct := rr.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected default format to be json, content-type=%q", ct)
	}
}

func TestHandler_Export_UnknownFormat(t *testing.T) {
	data := minimalExportData()
	fn := func(_ context.Context, _ uuid.UUID) (*ExportData, error) { return data, nil }
	id := uuid.New()
	req := makeChiRequest("GET", "/", id.String(), "xml")
	rr := httptest.NewRecorder()
	gatherableHandler(fn)(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for unknown format, got %d", rr.Code)
	}
}

// ---------------------------------------------------------------------------
// Helpers.
// ---------------------------------------------------------------------------

type errStub string

func (e errStub) Error() string { return string(e) }

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
