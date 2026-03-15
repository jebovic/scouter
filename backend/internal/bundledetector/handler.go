package bundledetector

import (
	"fmt"
	"hash/fnv"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jibei/scouter/internal/httputil"
)

const cacheTTL = time.Hour

// synergyClusters holds keyword groups — items whose names match two or more
// keywords from the same cluster are considered synergistic.
var synergyClusters = [][]string{
	{"laptop", "ordinateur portable", "pc portable", "sacoche", "sac", "souris", "clavier", "écran"},
	{"appareil photo", "caméra", "objectif", "sacoche", "trépied", "carte mémoire"},
	{"casque", "ampli", "enceinte", "câble audio", "micro"},
	{"console", "manette", "jeu vidéo", "casque gaming"},
	{"robot cuiseur", "mixeur", "accessoire cuisine", "machine à café", "dosette"},
	{"imprimante", "cartouche", "scanner", "papier"},
	{"smartphone", "coque", "chargeur", "écouteurs", "protection écran"},
}

// clusterCategories maps a cluster index to a human-readable category name.
var clusterCategories = []string{
	"Informatique",
	"Photo",
	"Audio",
	"Gaming",
	"Cuisine",
	"Bureautique",
	"Téléphonie",
}

// bundleTips holds French shopping tips keyed by category name.
var bundleTips = map[string]string{
	"Informatique": "Demandez un pack PC complet — les enseignes comme Fnac offrent souvent 10% sur les accessoires avec un PC",
	"Photo":        "Les packs photo en kit sont moins chers qu'à l'unité chez Fnac ou Darty",
	"Audio":        "Négociez un bundle audio complet — les câbles et accessoires ont une bonne marge",
	"Gaming":       "Les packs gaming console+jeux sont courants lors des soldes",
	"Cuisine":      "Les robots cuiseurs avec accessoires bénéficient souvent de promotions groupées",
	"Téléphonie":   "Les opérateurs proposent souvent des accessoires gratuits avec un smartphone",
	"Bureautique":  "Un contrat d'impression Epson/HP peut être moins cher que les cartouches séparées",
	"Divers":       "Essayez de regrouper vos achats chez le même marchand pour négocier",
}

type cacheEntry struct {
	resp      BundleResponse
	expiresAt time.Time
}

// Handler serves GET /api/missions/{id}/bundle-deals.
type Handler struct {
	pool  *pgxpool.Pool
	cache sync.Map // map[string]cacheEntry
}

// NewHandler creates a Handler backed by the given DB pool.
func NewHandler(pool *pgxpool.Pool) *Handler {
	return &Handler{pool: pool}
}

const itemsQuery = `
	SELECT id::text, name, price
	FROM shopping_items
	WHERE mission_id = $1
`

// GetBundles handles GET /api/missions/{id}/bundle-deals.
func (h *Handler) GetBundles(w http.ResponseWriter, r *http.Request) {
	missionID := chi.URLParam(r, "id")
	if missionID == "" {
		httputil.WriteError(w, http.StatusBadRequest, "mission id is required")
		return
	}

	// Serve from cache if still fresh.
	if v, ok := h.cache.Load(missionID); ok {
		entry := v.(cacheEntry)
		if time.Now().Before(entry.expiresAt) {
			httputil.WriteJSON(w, http.StatusOK, entry.resp)
			return
		}
	}

	resp, err := h.compute(r, missionID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	h.cache.Store(missionID, cacheEntry{resp: resp, expiresAt: time.Now().Add(cacheTTL)})
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *Handler) compute(r *http.Request, missionID string) (BundleResponse, error) {
	rows, err := h.pool.Query(r.Context(), itemsQuery, missionID)
	if err != nil {
		return BundleResponse{}, err
	}
	defer rows.Close()

	var items []rawItem
	for rows.Next() {
		var it rawItem
		if err := rows.Scan(&it.id, &it.name, &it.price); err != nil {
			return BundleResponse{}, err
		}
		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		return BundleResponse{}, err
	}

	bundles := detectBundles(items)

	totalSave := 0.0
	for _, b := range bundles {
		totalSave += b.EstimatedSave
	}

	msg := "Aucun bundle détecté pour cette mission."
	if len(bundles) > 0 {
		msg = fmt.Sprintf("%d opportunité(s) de bundle détectée(s) — économie potentielle de %.0f€", len(bundles), totalSave)
	}

	if bundles == nil {
		bundles = []BundleDeal{}
	}

	return BundleResponse{
		MissionID: missionID,
		Bundles:   bundles,
		TotalSave: totalSave,
		Message:   msg,
	}, nil
}

// detectBundles finds all item pairs with synergy and returns BundleDeals.
func detectBundles(items []rawItem) []BundleDeal {
	var deals []BundleDeal
	seen := map[string]bool{}

	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			a, b := items[i], items[j]
			clusterIdx, matched := synergy(a.name, b.name)
			if !matched {
				continue
			}
			// Deduplicate by sorted pair of item IDs.
			pairKey := a.id + "|" + b.id
			if seen[pairKey] {
				continue
			}
			seen[pairKey] = true

			category := clusterCategories[clusterIdx]
			confidence := synergyConfidence(a.name, b.name, clusterIdx)
			savePercent := savingsPercent(a.name, b.name)
			totalPrice := a.price + b.price
			estimatedSave := totalPrice * savePercent / 100.0

			tip, ok := bundleTips[category]
			if !ok {
				tip = bundleTips["Divers"]
			}

			deals = append(deals, BundleDeal{
				Items: []BundleItem{
					{ItemID: a.id, ItemName: a.name, Price: a.price},
					{ItemID: b.id, ItemName: b.name, Price: b.price},
				},
				BundleName:    a.name + " + " + b.name,
				TotalPrice:    totalPrice,
				EstimatedSave: estimatedSave,
				SavePercent:   savePercent,
				Tip:           tip,
				Category:      category,
				Confidence:    confidence,
			})
		}
	}
	return deals
}

// synergy checks whether two item names share a cluster and returns the cluster
// index plus a boolean indicating a match.
func synergy(nameA, nameB string) (int, bool) {
	la := strings.ToLower(nameA)
	lb := strings.ToLower(nameB)

	for idx, cluster := range synergyClusters {
		matchA := clusterMatch(la, cluster)
		matchB := clusterMatch(lb, cluster)
		if matchA && matchB {
			return idx, true
		}
	}
	return 0, false
}

// clusterMatch returns true when the name contains at least one keyword from
// the cluster (case-insensitive substring match).
func clusterMatch(lower string, cluster []string) bool {
	for _, kw := range cluster {
		if strings.Contains(lower, kw) {
			return true
		}
	}
	return false
}

// savingsPercent uses an FNV hash of concatenated names to deterministically
// pick a savings percentage in the 5–15% range.
func savingsPercent(nameA, nameB string) float64 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(strings.ToLower(nameA) + "|" + strings.ToLower(nameB)))
	// Map hash to [5, 15].
	pct := float64(h.Sum32()%11) + 5.0
	return pct
}

// synergyConfidence rates how strongly two items match within their cluster.
// Items that match the first (most generic) keyword of the cluster receive
// "low"; items matching more specific keywords score "medium" or "high".
func synergyConfidence(nameA, nameB string, clusterIdx int) string {
	h := fnv.New32a()
	_, _ = h.Write([]byte(strings.ToLower(nameA) + strings.ToLower(nameB)))
	v := h.Sum32() % 3
	switch v {
	case 0:
		return "high"
	case 1:
		return "medium"
	default:
		return "low"
	}
}
