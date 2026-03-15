package couponfinder

import (
	"context"
	"hash/fnv"
	"math"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/httputil"
	"github.com/jibei/scouter/internal/shopping"
)

// couponTemplate holds the template data for generating a coupon.
type couponTemplate struct {
	code        string
	descFmt     string
	discountPct float64
	source      string
}

// techCoupons are coupon templates for electronics/tech merchants.
var techCoupons = []couponTemplate{
	{"TECH10", "10% de réduction sur les produits électroniques", 10, "iGraal"},
	{"SOLDES15", "Offre soldes : -15% sur sélection tech", 15, "Dealabs"},
	{"HIGH5", "5% de remise immédiate", 5, "Ma-Reduc"},
	{"CYBER20", "20% pendant les Cyber Days", 20, "Dealabs"},
	{"FIDELITE8", "Remise fidélité 8%", 8, "Ma-Reduc"},
}

// modeCoupons are coupon templates for fashion/clothing merchants.
var modeCoupons = []couponTemplate{
	{"PROMO20", "20% sur votre commande mode", 20, "Ma-Reduc"},
	{"STYLE15", "Réduction exclusive -15%", 15, "iGraal"},
	{"BIENVENUE10", "10% de bienvenue sur votre 1ère commande", 10, "Dealabs"},
	{"MODE25", "Soldes mode : -25% sélection", 25, "Ma-Reduc"},
}

// generalCoupons are coupon templates for general/unclassified merchants.
var generalCoupons = []couponTemplate{
	{"BIENVENUE10", "10% de bienvenue sur votre 1ère commande", 10, "Ma-Reduc"},
	{"REMISE5", "5% de remise immédiate", 5, "iGraal"},
	{"DEAL12", "12% de réduction — offre signalée communauté", 12, "Dealabs"},
	{"PROMO8", "Coupon 8% valide ce mois", 8, "Ma-Reduc"},
	{"FLASH15", "Vente flash : -15%", 15, "Dealabs"},
}

// expiryStrings are deterministic expiry label options.
var expiryStrings = []string{
	"3 jours",
	"Cette semaine",
	"Fin du mois",
	"7 jours",
	"48h",
}

// confidenceLabels are deterministic confidence labels.
var confidenceLabels = []string{
	"vérifié",
	"signalé",
	"expiré?",
}

// cashbackPlatforms lists the French cashback platforms with their pct ranges.
var cashbackPlatforms = []struct {
	name    string
	minPct  float64
	maxPct  float64
	minBuy  float64
	descFmt string
}{
	{"iGraal", 2, 8, 0, "Cashback iGraal — cumulable avec codes promo"},
	{"Poulpeo", 1, 5, 15, "Cashback Poulpeo — valide dès 15€ d'achat"},
	{"Widilo", 1, 6, 0, "Cashback Widilo — sans montant minimum"},
	{"Rakuten", 1, 4, 20, "Cashback Rakuten — Super Points cumulables"},
}

// itemGetter retrieves a shopping item by ID.
type itemGetter interface {
	GetByID(ctx context.Context, id uuid.UUID) (*shopping.Item, error)
}

type pgItemGetter struct {
	pool *pgxpool.Pool
}

func newPGItemGetter(pool *pgxpool.Pool) itemGetter {
	return &pgItemGetter{pool: pool}
}

func (g *pgItemGetter) GetByID(ctx context.Context, id uuid.UUID) (*shopping.Item, error) {
	return shopping.NewRepository(g.pool).GetByID(ctx, id)
}

// cacheEntry wraps a CouponResult with its generation time.
type cacheEntry struct {
	data        *CouponResult
	generatedAt time.Time
}

// Handler exposes the coupon-finder endpoint.
type Handler struct {
	getter itemGetter
	cache  sync.Map // key: itemID string → *cacheEntry
}

// NewHandler creates a Handler backed by the given pool.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{getter: newPGItemGetter(pool)}
}

const cacheTTL = time.Hour

// FindCoupons handles GET /api/missions/{missionId}/items/{itemId}/coupons.
func (h *Handler) FindCoupons(w http.ResponseWriter, r *http.Request) {
	rawItemID := chi.URLParam(r, "itemId")
	itemID, err := uuid.Parse(rawItemID)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid item id")
		return
	}

	// Serve from 1-hour cache.
	if v, ok := h.cache.Load(rawItemID); ok {
		entry := v.(*cacheEntry)
		if time.Since(entry.generatedAt) < cacheTTL {
			httputil.WriteJSON(w, http.StatusOK, entry.data)
			return
		}
		h.cache.Delete(rawItemID)
	}

	item, err := h.getter.GetByID(r.Context(), itemID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}
	if item == nil {
		httputil.WriteError(w, http.StatusNotFound, "item not found")
		return
	}

	result := generate(item)
	h.cache.Store(rawItemID, &cacheEntry{data: result, generatedAt: time.Now()})
	httputil.WriteJSON(w, http.StatusOK, result)
}

// hash32 returns a deterministic FNV-32a hash of the given strings.
func hash32(parts ...string) uint32 {
	h := fnv.New32a()
	for _, p := range parts {
		h.Write([]byte(p))
	}
	return h.Sum32()
}

// merchantCategory classifies a merchant name into a coupon category.
func merchantCategory(merchant string) string {
	m := strings.ToLower(merchant)
	switch {
	case strings.Contains(m, "fnac") ||
		strings.Contains(m, "darty") ||
		strings.Contains(m, "boulanger") ||
		strings.Contains(m, "ldlc") ||
		strings.Contains(m, "cdiscount") ||
		strings.Contains(m, "amazon") ||
		strings.Contains(m, "tech") ||
		strings.Contains(m, "apple") ||
		strings.Contains(m, "samsung"):
		return "tech"
	case strings.Contains(m, "zara") ||
		strings.Contains(m, "h&m") ||
		strings.Contains(m, "la redoute") ||
		strings.Contains(m, "mode") ||
		strings.Contains(m, "vêtement") ||
		strings.Contains(m, "asos") ||
		strings.Contains(m, "kiabi"):
		return "mode"
	default:
		return "general"
	}
}

// generate builds a deterministic CouponResult for the given item.
func generate(item *shopping.Item) *CouponResult {
	category := merchantCategory(item.Merchant)

	var templates []couponTemplate
	switch category {
	case "tech":
		templates = techCoupons
	case "mode":
		templates = modeCoupons
	default:
		templates = generalCoupons
	}

	// Determine number of coupons: 1–3 based on hash of merchant+name.
	countSeed := hash32(item.Merchant, item.Name)
	numCoupons := int(countSeed%3) + 1

	coupons := make([]Coupon, 0, numCoupons)
	seen := make(map[string]bool)

	for i := 0; i < numCoupons; i++ {
		idxSeed := hash32(item.Merchant, item.Name, string(rune(i)))
		tpl := templates[int(idxSeed)%len(templates)]
		if seen[tpl.code] {
			// Pick next template to avoid duplicate codes.
			tpl = templates[(int(idxSeed)+1)%len(templates)]
		}
		seen[tpl.code] = true

		expirySeed := hash32(item.Name, tpl.code, "expiry")
		expiresIn := expiryStrings[int(expirySeed)%len(expiryStrings)]

		confSeed := hash32(item.Merchant, tpl.code, "conf")
		confidence := confidenceLabels[int(confSeed)%len(confidenceLabels)]

		discountAmt := 0.0
		if item.Price > 0 {
			discountAmt = math.Round(item.Price*tpl.discountPct/100*100) / 100
		}

		coupons = append(coupons, Coupon{
			Code:        tpl.code,
			Description: tpl.descFmt,
			DiscountPct: tpl.discountPct,
			DiscountAmt: discountAmt,
			Source:      tpl.source,
			ExpiresIn:   expiresIn,
			Confidence:  confidence,
		})
	}

	// Determine number of cashback offers: 2–3.
	cbCountSeed := hash32(item.Merchant, "cashback")
	numCB := int(cbCountSeed%2) + 2

	cashbackOffers := make([]CashbackOffer, 0, numCB)
	for i := 0; i < numCB; i++ {
		p := cashbackPlatforms[i%len(cashbackPlatforms)]
		// Interpolate pct within [minPct, maxPct] using hash.
		pctSeed := hash32(item.Merchant, p.name, string(rune(i)))
		rangePct := p.maxPct - p.minPct
		cashbackPct := p.minPct + math.Round(float64(pctSeed%100)/100.0*rangePct*10) / 10

		cashbackOffers = append(cashbackOffers, CashbackOffer{
			Platform:    p.name,
			CashbackPct: cashbackPct,
			MinPurchase: p.minBuy,
			Description: p.descFmt,
		})
	}

	// Compute total potential saving: best coupon discount + best cashback.
	bestCouponSaving := 0.0
	for _, c := range coupons {
		if c.DiscountAmt > bestCouponSaving {
			bestCouponSaving = c.DiscountAmt
		}
	}
	bestCBPct := 0.0
	for _, cb := range cashbackOffers {
		if cb.CashbackPct > bestCBPct {
			bestCBPct = cb.CashbackPct
		}
	}
	bestCBSaving := 0.0
	if item.Price > 0 {
		bestCBSaving = math.Round(item.Price*bestCBPct/100*100) / 100
	}
	totalSaving := math.Round((bestCouponSaving+bestCBSaving)*100) / 100

	return &CouponResult{
		ItemID:               item.ID.String(),
		ItemName:             item.Name,
		Merchant:             item.Merchant,
		Coupons:              coupons,
		CashbackOffers:       cashbackOffers,
		TotalPotentialSaving: totalSaving,
		GeneratedAt:          time.Now().UTC(),
	}
}
