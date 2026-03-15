package frenchmarket

import "strings"

// categoryProfile holds the static market profile for a product category.
type categoryProfile struct {
	retailers          []RetailerPricing
	bestMonthToBuy     string
	flashSaleFrequency string
	tips               []string
	marketTrend        string
	priceMin           float64
	priceMax           float64
	avgPrice           float64
}

// knowledgeBase maps normalised category keywords to their French market profile.
var knowledgeBase = map[string]categoryProfile{
	"électronique": {
		retailers: []RetailerPricing{
			{Name: "Fnac", URL: "https://www.fnac.com", TypicalDiscount: "5–25 %", LoyaltyProgram: "Fnac+"},
			{Name: "Darty", URL: "https://www.darty.com", TypicalDiscount: "5–20 %", LoyaltyProgram: "Darty Max"},
			{Name: "Boulanger", URL: "https://www.boulanger.com", TypicalDiscount: "10–30 %", LoyaltyProgram: "Boulanger Fidélité"},
			{Name: "Amazon.fr", URL: "https://www.amazon.fr", TypicalDiscount: "10–40 %", LoyaltyProgram: "Amazon Prime"},
			{Name: "Cdiscount", URL: "https://www.cdiscount.com", TypicalDiscount: "15–50 %", LoyaltyProgram: "Cdiscount à volonté"},
		},
		bestMonthToBuy:     "novembre (Black Friday & Cyber Monday)",
		flashSaleFrequency: "fréquentes (plusieurs fois par semaine)",
		tips: []string{
			"Activez les alertes prix sur Idealo.fr pour suivre les variations.",
			"Comparez toujours entre Fnac, Darty et Boulanger — leurs prix diffèrent souvent de 5–15 %.",
			"Les garanties constructeur sont de 2 ans minimum en France (directive UE).",
		},
		marketTrend: "stable",
		priceMin:    50,
		priceMax:    3000,
		avgPrice:    400,
	},
	"électroménager": {
		retailers: []RetailerPricing{
			{Name: "Boulanger", URL: "https://www.boulanger.com", TypicalDiscount: "10–30 %", LoyaltyProgram: "Boulanger Fidélité"},
			{Name: "Darty", URL: "https://www.darty.com", TypicalDiscount: "10–25 %", LoyaltyProgram: "Darty Max"},
			{Name: "Fnac", URL: "https://www.fnac.com", TypicalDiscount: "5–20 %", LoyaltyProgram: "Fnac+"},
			{Name: "Conforama", URL: "https://www.conforama.fr", TypicalDiscount: "20–40 %", LoyaltyProgram: "Carte Conforama"},
		},
		bestMonthToBuy:     "janvier (soldes d'hiver) et juin (soldes d'été)",
		flashSaleFrequency: "régulières (ventes flash hebdomadaires)",
		tips: []string{
			"Profitez des soldes légales (janvier et juillet) pour les achats importants.",
			"L'étiquette énergétique EU est obligatoire — privilégiez A ou A+ pour économiser sur la facture.",
			"Certains appareils sont éligibles à la prime énergie ou à MaPrimeRénov'.",
		},
		marketTrend: "en légère hausse (matières premières)",
		priceMin:    100,
		priceMax:    5000,
		avgPrice:    600,
	},
	"mode": {
		retailers: []RetailerPricing{
			{Name: "Zara", URL: "https://www.zara.com/fr", TypicalDiscount: "20–50 %", LoyaltyProgram: "Zara Membership"},
			{Name: "H&M", URL: "https://www.hm.com/fr_fr", TypicalDiscount: "20–50 %", LoyaltyProgram: "H&M Member"},
			{Name: "ASOS", URL: "https://www.asos.com/fr", TypicalDiscount: "20–60 %", LoyaltyProgram: "ASOS Premier"},
			{Name: "La Redoute", URL: "https://www.laredoute.fr", TypicalDiscount: "30–70 %", LoyaltyProgram: "La Redoute+"},
			{Name: "Kiabi", URL: "https://www.kiabi.com", TypicalDiscount: "30–60 %", LoyaltyProgram: "Kiabi Money"},
		},
		bestMonthToBuy:     "janvier et juillet (soldes légales de 4 semaines)",
		flashSaleFrequency: "fréquentes (ventes privées Veepee, Showroomprivé)",
		tips: []string{
			"Les soldes légales durent 4 semaines en France — les remises atteignent souvent 70 % en fin de période.",
			"Inscrivez-vous à Veepee et Showroomprivé pour les ventes privées hors soldes.",
			"Vérifiez les tailles européennes : la taille FR peut différer du pays d'origine.",
		},
		marketTrend: "stable avec montée du seconde main",
		priceMin:    10,
		priceMax:    500,
		avgPrice:    60,
	},
	"alimentation": {
		retailers: []RetailerPricing{
			{Name: "E.Leclerc", URL: "https://www.leclerc.fr", TypicalDiscount: "5–30 %", LoyaltyProgram: "Carte E.Leclerc"},
			{Name: "Carrefour", URL: "https://www.carrefour.fr", TypicalDiscount: "5–25 %", LoyaltyProgram: "Carte Carrefour+"},
			{Name: "Intermarché", URL: "https://www.intermarche.com", TypicalDiscount: "10–30 %", LoyaltyProgram: "Carte Intermarché"},
			{Name: "Monoprix", URL: "https://www.monoprix.fr", TypicalDiscount: "5–20 %", LoyaltyProgram: "Carte Monoprix"},
		},
		bestMonthToBuy:     "toute l'année (bonnes affaires permanentes)",
		flashSaleFrequency: "quotidiennes (promotions tournantes)",
		tips: []string{
			"Comparez avec l'application Quiestlemoinscher.com avant tout achat en grande surface.",
			"Les MDD (marques distributeur) offrent un rapport qualité/prix souvent supérieur aux marques nationales.",
			"Planifiez vos achats volumineux pendant les opérations 'Prix bloqués' de Leclerc.",
		},
		marketTrend: "inflation persistante (+3–5 % par an)",
		priceMin:    1,
		priceMax:    200,
		avgPrice:    30,
	},
	"voyage": {
		retailers: []RetailerPricing{
			{Name: "SNCF Connect", URL: "https://www.sncf-connect.com", TypicalDiscount: "20–60 %", LoyaltyProgram: "Programme Avantage"},
			{Name: "BlaBlaCar", URL: "https://www.blablacar.fr", TypicalDiscount: "50–80 %", LoyaltyProgram: "BlaBlaCar+"},
			{Name: "Ouigo", URL: "https://www.ouigo.com", TypicalDiscount: "30–70 %", LoyaltyProgram: "–"},
			{Name: "Flixbus", URL: "https://www.flixbus.fr", TypicalDiscount: "20–60 %", LoyaltyProgram: "Flix Promo"},
		},
		bestMonthToBuy:     "3 mois à l'avance (tarifs Prem's SNCF)",
		flashSaleFrequency: "régulières (flash SNCF le mardi matin)",
		tips: []string{
			"Réservez les billets SNCF dès l'ouverture des Prem's (90 jours avant) pour les prix les plus bas.",
			"La carte Avantage Jeune (-26 ans) offre jusqu'à 60 % de réduction sur les TGV.",
			"Combinez Ouigo + BlaBlaCar pour les trajets péri-urbains non desservis.",
		},
		marketTrend: "en hausse (grèves et demande post-Covid)",
		priceMin:    5,
		priceMax:    500,
		avgPrice:    80,
	},
	"sport": {
		retailers: []RetailerPricing{
			{Name: "Decathlon", URL: "https://www.decathlon.fr", TypicalDiscount: "10–50 %", LoyaltyProgram: "Carte Decathlon"},
			{Name: "Go Sport", URL: "https://www.go-sport.com", TypicalDiscount: "10–40 %", LoyaltyProgram: "Club Go Sport"},
			{Name: "Amazon.fr", URL: "https://www.amazon.fr", TypicalDiscount: "10–40 %", LoyaltyProgram: "Amazon Prime"},
		},
		bestMonthToBuy:     "janvier (bonnes résolutions + soldes)",
		flashSaleFrequency: "modérées (promotions saisonnières)",
		tips: []string{
			"Decathlon reste la référence rapport qualité/prix pour le sport en France.",
			"Les fins de saison (mars pour le ski, septembre pour l'été) offrent les meilleures remises.",
			"Vérifiez les articles reconditionnés sur le site de Decathlon.",
		},
		marketTrend: "stable",
		priceMin:    10,
		priceMax:    2000,
		avgPrice:    150,
	},
}

// categoryAliases maps common alternative spellings / English terms to canonical keys.
var categoryAliases = map[string]string{
	"electronics":  "électronique",
	"tech":         "électronique",
	"informatique": "électronique",
	"hi-fi":        "électronique",
	"hifi":         "électronique",
	"vêtements":    "mode",
	"vetements":    "mode",
	"textile":      "mode",
	"clothes":      "mode",
	"fashion":      "mode",
	"épicerie":     "alimentation",
	"epicerie":     "alimentation",
	"food":         "alimentation",
	"nourriture":   "alimentation",
	"transport":    "voyage",
	"travel":       "voyage",
	"train":        "voyage",
	"fitness":      "sport",
	"gym":          "sport",
}

// defaultProfile is returned when no category matches.
var defaultProfile = categoryProfile{
	retailers: []RetailerPricing{
		{Name: "Amazon.fr", URL: "https://www.amazon.fr", TypicalDiscount: "10–40 %", LoyaltyProgram: "Amazon Prime"},
		{Name: "Cdiscount", URL: "https://www.cdiscount.com", TypicalDiscount: "15–50 %", LoyaltyProgram: "Cdiscount à volonté"},
		{Name: "Fnac", URL: "https://www.fnac.com", TypicalDiscount: "5–25 %", LoyaltyProgram: "Fnac+"},
	},
	bestMonthToBuy:     "novembre (Black Friday) et janvier/juillet (soldes légales)",
	flashSaleFrequency: "variables selon la catégorie",
	tips: []string{
		"Utilisez Idealo.fr ou LeGuide.com pour comparer les prix entre marchands français.",
		"Les soldes légales (janvier et juillet) offrent des remises officielles dans tout le commerce français.",
		"Vérifiez que le vendeur est établi en France pour bénéficier de la garantie légale de 2 ans.",
	},
	marketTrend: "variable",
	priceMin:    20,
	priceMax:    500,
	avgPrice:    100,
}

// GetInsight returns French market intelligence for the given item and category.
// The category lookup is case-insensitive and supports common aliases.
func GetInsight(itemName, category string) MarketInsight {
	profile := resolveProfile(category)
	retailers := profile.retailers
	if len(retailers) > 3 {
		retailers = retailers[:3]
	}

	return MarketInsight{
		ItemName:            itemName,
		Category:            category,
		AveragePriceEUR:     profile.avgPrice,
		PriceRange:          PriceRange{Min: profile.priceMin, Max: profile.priceMax},
		BestRetailers:       retailers,
		BestMonthToBuy:      profile.bestMonthToBuy,
		FlashSaleFrequency:  profile.flashSaleFrequency,
		TipsForFrenchBuyers: profile.tips,
		MarketTrend:         profile.marketTrend,
	}
}

// resolveProfile looks up the most specific profile for the given category string.
func resolveProfile(category string) categoryProfile {
	lower := strings.ToLower(strings.TrimSpace(category))
	if lower == "" {
		return defaultProfile
	}

	// Direct match.
	if p, ok := knowledgeBase[lower]; ok {
		return p
	}

	// Alias match.
	if canonical, ok := categoryAliases[lower]; ok {
		if p, ok := knowledgeBase[canonical]; ok {
			return p
		}
	}

	// Partial substring match against known keys and aliases.
	for key, p := range knowledgeBase {
		if strings.Contains(lower, key) || strings.Contains(key, lower) {
			return p
		}
	}
	for alias, canonical := range categoryAliases {
		if strings.Contains(lower, alias) {
			if p, ok := knowledgeBase[canonical]; ok {
				return p
			}
		}
	}

	return defaultProfile
}
