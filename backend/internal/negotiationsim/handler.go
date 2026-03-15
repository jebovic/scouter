package negotiationsim

import (
	"context"
	"fmt"
	"hash/fnv"
	"math"
	"net/http"
	"sync"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jibei/scouter/internal/httputil"
	"github.com/jibei/scouter/internal/shopping"
)

// itemGetter retrieves a shopping item by UUID.
type itemGetter interface {
	GetByID(ctx context.Context, id uuid.UUID) (*shopping.Item, error)
}

// Handler exposes the negotiation script endpoint over HTTP.
type Handler struct {
	getter itemGetter
	cache  sync.Map // key: itemID string → *NegotiationScript
}

// NewHandler creates a Handler backed by the given shopping repository.
func NewHandler(getter itemGetter) *Handler {
	return &Handler{getter: getter}
}

// GetScript handles GET /api/missions/{missionId}/items/{itemId}/negotiation-script.
// Returns a deterministic NegotiationScript for the item; results are cached for 1h.
func (h *Handler) GetScript(w http.ResponseWriter, r *http.Request) {
	rawItemID := chi.URLParam(r, "itemId")
	itemID, err := uuid.Parse(rawItemID)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid item id")
		return
	}

	// Serve from 1h in-memory cache.
	if v, ok := h.cache.Load(rawItemID); ok {
		script := v.(*NegotiationScript)
		if time.Since(script.GeneratedAt) < time.Hour {
			httputil.WriteJSON(w, http.StatusOK, script)
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

	script := generate(item)
	h.cache.Store(rawItemID, script)
	httputil.WriteJSON(w, http.StatusOK, script)
}

// generate builds a deterministic NegotiationScript from an item.
// The FNV-32a hash of the item ID selects among pre-written script templates so
// the result is stable across requests without an LLM call.
func generate(item *shopping.Item) *NegotiationScript {
	targetPrice := item.Price * 0.85
	if item.TargetPrice != nil {
		targetPrice = *item.TargetPrice
	}
	savings := math.Round((item.Price-targetPrice)*100) / 100
	if savings < 0 {
		savings = 0
	}

	h := fnv.New32a()
	h.Write([]byte(item.ID.String()))
	seed := int(h.Sum32())

	steps := buildSteps(item.Name, item.Price, targetPrice, seed)
	tips := buildTips(seed)

	return &NegotiationScript{
		ItemID:           item.ID.String(),
		ItemName:         item.Name,
		CurrentPrice:     item.Price,
		TargetPrice:      targetPrice,
		PotentialSavings: savings,
		Scripts:          steps,
		Tips:             tips,
		GeneratedAt:      time.Now().UTC(),
	}
}

// openingTemplates contains French opening lines. {item} is replaced at runtime.
var openingTemplates = []string{
	"Bonjour, je suis intéressé par %s, mais je voulais vérifier si vous pouviez faire un geste sur le prix.",
	"Bonjour, votre %s m'intéresse vraiment. Est-ce que vous avez une marge sur le tarif affiché ?",
	"Bonjour, j'ai vu le %s sur votre site. Est-il possible d'obtenir un meilleur prix si j'achète aujourd'hui ?",
}

var openingVendorTemplates = []string{
	"Bonjour ! Oui, c'est un très bon produit. Quel budget avez-vous en tête ?",
	"Bonjour ! Je vais voir ce que je peux faire. Vous êtes prêt à acheter aujourd'hui ?",
	"Bonjour ! Effectivement, c'est notre meilleur modèle. Quelle est votre fourchette ?",
}

var openingTips = []string{
	"Soyez souriant et détendu — la première impression est déterminante.",
	"Ne révélez pas encore votre budget maximum.",
	"Montrez de l'intérêt sans paraître trop pressé d'acheter.",
}

// negotiationTemplates contain the key negotiation exchange.
var negotiationTemplates = []string{
	"J'ai vu ce produit à %.0f€ ailleurs, pouvez-vous vous aligner ?",
	"Mon budget est limité à %.0f€. Pouvez-vous faire un effort ?",
	"Je suis prêt à acheter immédiatement si vous descendez à %.0f€.",
}

var negotiationVendorTemplates = []string{
	"Je comprends, laissez-moi vérifier ce que je peux faire pour vous.",
	"C'est un peu juste pour nous, mais je vais parler à mon responsable.",
	"Honnêtement, c'est notre meilleur prix, mais je peux peut-être inclure quelque chose en plus.",
}

var negotiationTips = []string{
	"Mentionnez la concurrence sans être agressif — restez factuel.",
	"Gardez le silence après votre offre : laissez le vendeur répondre.",
	"Ne donnez pas deux chiffres — une offre précise est plus crédible.",
}

// closingTemplates contain the closing line with target price.
var closingTemplates = []string{
	"Si vous pouvez baisser à %.0f€, je l'achète immédiatement.",
	"Faites-moi un prix à %.0f€ et c'est conclu tout de suite.",
	"Je signe maintenant si vous arrivez à %.0f€ — c'est ma dernière offre.",
}

var closingVendorTemplates = []string{
	"D'accord, je vais faire un effort. On peut s'entendre à ce prix.",
	"C'est limite, mais je veux que vous soyez satisfait. Marché conclu.",
	"OK, je valide avec mon manager et je reviens vers vous dans deux minutes.",
}

var closingTips = []string{
	"Proposez un paiement immédiat comme levier — c'est souvent décisif.",
	"Ne montrez pas votre enthousiasme trop tôt : restez neutre jusqu'à la signature.",
	"Prévoyez une concession de rechange (accessoire gratuit, livraison offerte) si le prix est bloqué.",
}

// staticTips are general French negotiation tips shown in the tips section.
var staticTips = [][]string{
	{
		"Arrivez avec une offre précise, pas une fourchette.",
		"Mentionnez la concurrence sans être agressif.",
		"Proposez un paiement immédiat comme levier.",
		"Restez calme et souriant tout au long de la négociation.",
	},
	{
		"Renseignez-vous sur le prix moyen du marché avant de négocier.",
		"Choisissez le bon moment : fin de mois, fin de saison, ou après les soldes.",
		"Ne négociez jamais par e-mail — faites-le en face à face.",
		"Soyez prêt à partir si l'offre ne vous convient pas.",
	},
	{
		"Utilisez le silence à votre avantage : laissez le vendeur parler.",
		"Demandez des extras si le prix est fixe : garantie prolongée, livraison offerte.",
		"Mentionnez que vous comparez avec un autre vendeur.",
		"Formulez toujours votre offre de manière positive.",
	},
}

func pick(templates []string, seed, phase int) string {
	return templates[(seed+phase)%len(templates)]
}

func buildSteps(itemName string, currentPrice, targetPrice float64, seed int) []ScriptStep {
	return []ScriptStep{
		{
			Phase:   "opening",
			Speaker: "vous",
			Text:    fmt.Sprintf(pick(openingTemplates, seed, 0), itemName),
			Tip:     pick(openingTips, seed, 0),
		},
		{
			Phase:   "opening",
			Speaker: "vendeur",
			Text:    pick(openingVendorTemplates, seed, 1),
			Tip:     "",
		},
		{
			Phase:   "negotiation",
			Speaker: "vous",
			Text:    fmt.Sprintf(pick(negotiationTemplates, seed, 2), targetPrice),
			Tip:     pick(negotiationTips, seed, 2),
		},
		{
			Phase:   "negotiation",
			Speaker: "vendeur",
			Text:    pick(negotiationVendorTemplates, seed, 3),
			Tip:     "",
		},
		{
			Phase:   "closing",
			Speaker: "vous",
			Text:    fmt.Sprintf(pick(closingTemplates, seed, 4), targetPrice),
			Tip:     pick(closingTips, seed, 4),
		},
		{
			Phase:   "closing",
			Speaker: "vendeur",
			Text:    pick(closingVendorTemplates, seed, 5),
			Tip:     "",
		},
	}
}

func buildTips(seed int) []string {
	return staticTips[seed%len(staticTips)]
}
