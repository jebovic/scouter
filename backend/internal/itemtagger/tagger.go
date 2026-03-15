package itemtagger

import (
	"strings"
	"time"
)

// categoryRule maps a set of keywords to a category name.
type categoryRule struct {
	keywords []string
	category string
}

var categoryRules = []categoryRule{
	{keywords: []string{"iphone", "samsung galaxy", "pixel", "smartphone", "téléphone", "telephone", "oneplus", "redmi", "xiaomi phone", "oppo", "realme"}, category: "Téléphonie"},
	{keywords: []string{"macbook", "laptop", "pc portable", "lenovo", "thinkpad", "dell xps", "asus zenbook", "hp spectre", "surface pro", "chromebook"}, category: "Informatique"},
	{keywords: []string{"ipad", "tablette", "tablet", "galaxy tab", "surface go"}, category: "Tablettes"},
	{keywords: []string{"tv ", "télévision", "television", "oled", "qled", "neo qled", "miniled", "4k tv", "8k tv", "smart tv"}, category: "TV & Audio"},
	{keywords: []string{"casque", "écouteurs", "airpods", "galaxy buds", "soundbar", "enceinte", "speaker", "audiophile", "bose", "sennheiser", "jbl"}, category: "Audio"},
	{keywords: []string{"aspirateur", "robot ménager", "lave-linge", "lave linge", "sèche-linge", "seche linge", "réfrigérateur", "refrigerateur", "lave-vaisselle", "micro-onde", "microwave", "mixeur", "cafetière", "machine à café", "machine cafe"}, category: "Électroménager"},
	{keywords: []string{"dyson", "roomba", "irobot", "aspirateur balai", "robot aspirateur"}, category: "Électroménager"},
	{keywords: []string{"chaussures", "sneakers", "nike", "adidas", "jordan", "baskets", "boots", "running", "trail"}, category: "Mode"},
	{keywords: []string{"vêtement", "vetement", "chemise", "pantalon", "robe", "manteau", "veste", "hoodie", "pull"}, category: "Mode"},
	{keywords: []string{"sac", "sac à main", "handbag", "backpack", "sac à dos", "luggage", "valise", "trolley"}, category: "Maroquinerie"},
	{keywords: []string{"montre", "watch", "horloge", "smartwatch", "apple watch", "garmin", "fitbit"}, category: "Montres & Bijoux"},
	{keywords: []string{"bijou", "collier", "bracelet", "bague", "ring", "necklace", "earring", "boucle d'oreille"}, category: "Montres & Bijoux"},
	{keywords: []string{"console", "playstation", "ps5", "xbox", "nintendo", "switch", "gaming", "jeu vidéo", "jeu video", "manette", "controller"}, category: "Gaming"},
	{keywords: []string{"appareil photo", "camera", "objectif", "lens", "reflex", "hybride", "gopro", "drone"}, category: "Photo & Vidéo"},
	{keywords: []string{"vélo", "velo", "trottinette", "scooter électrique", "ebike", "e-bike", "cyclisme"}, category: "Mobilité Douce"},
	{keywords: []string{"voiture", "auto", "automobile", "pneu", "accessoire auto"}, category: "Automobile"},
	{keywords: []string{"mobilier", "meuble", "canapé", "canape", "table", "chaise", "lit", "matelas", "armoire", "bureau"}, category: "Maison & Mobilier"},
	{keywords: []string{"livre", "roman", "bande dessinée", "bd", "manga", "ebook", "audiobook"}, category: "Livres & Culture"},
	{keywords: []string{"voyage", "billet d'avion", "hotel", "hôtel", "airbnb", "séjour", "sejour", "croisière", "croisiere"}, category: "Voyage"},
	{keywords: []string{"sport", "fitness", "musculation", "yoga", "vélo stationnaire", "tapis roulant", "rameur", "haltère"}, category: "Sport & Fitness"},
	{keywords: []string{"parfum", "cosmétique", "cosmetique", "maquillage", "crème", "creme", "sérum", "serum", "shampoing"}, category: "Beauté & Soins"},
	{keywords: []string{"alimentation", "épicerie", "epicerie", "bio", "supplement", "compléments alimentaires", "protéine", "whey"}, category: "Alimentation"},
	{keywords: []string{"jouet", "lego", "puzzle", "peluche", "doudou", "figurine", "jeu de société"}, category: "Jouets & Enfants"},
}

var brands = []string{
	// Tech
	"apple", "samsung", "sony", "lg", "microsoft", "google", "huawei", "xiaomi", "oppo",
	"oneplus", "motorola", "nokia", "lenovo", "dell", "hp", "asus", "acer", "msi", "razer",
	"bose", "jbl", "sennheiser", "audio-technica", "beats", "harman", "yamaha", "pioneer",
	"canon", "nikon", "fujifilm", "gopro", "dji",
	// Home appliances
	"dyson", "philips", "bosch", "siemens", "miele", "whirlpool", "electrolux", "tefal",
	"moulinex", "kitchenaid", "nespresso", "de'longhi", "delonghi", "irobot", "roomba",
	// Fashion
	"nike", "adidas", "puma", "new balance", "reebok", "under armour", "vans", "converse",
	// Luxury
	"louis vuitton", "chanel", "hermès", "hermes", "gucci", "prada", "dior", "burberry",
	"rolex", "omega", "cartier", "tag heuer", "patek philippe", "audemars piguet", "breitling",
	"porsche", "ferrari", "lamborghini", "bentley", "maserati",
	// Gaming
	"nintendo", "playstation", "xbox",
}

var luxuryBrands = map[string]bool{
	"louis vuitton":   true,
	"chanel":          true,
	"hermès":          true,
	"hermes":          true,
	"gucci":           true,
	"prada":           true,
	"dior":            true,
	"burberry":        true,
	"rolex":           true,
	"omega":           true,
	"cartier":         true,
	"tag heuer":       true,
	"patek philippe":  true,
	"audemars piguet": true,
	"breitling":       true,
	"porsche":         true,
	"ferrari":         true,
	"lamborghini":     true,
	"bentley":         true,
	"maserati":        true,
}

var techCategories = map[string]bool{
	"Téléphonie":  true,
	"Informatique": true,
	"Tablettes":   true,
	"TV & Audio":  true,
	"Audio":       true,
	"Gaming":      true,
	"Photo & Vidéo": true,
}

// productTypeMap maps keywords to a product type label.
var productTypeKeywords = []struct {
	keyword     string
	productType string
}{
	{"iphone", "smartphone"},
	{"smartphone", "smartphone"},
	{"téléphone", "smartphone"},
	{"macbook", "laptop"},
	{"laptop", "laptop"},
	{"pc portable", "laptop"},
	{"chromebook", "laptop"},
	{"ipad", "tablette"},
	{"tablette", "tablette"},
	{"tablet", "tablette"},
	{"tv ", "télévision"},
	{"télévision", "télévision"},
	{"television", "télévision"},
	{"oled", "télévision"},
	{"qled", "télévision"},
	{"casque", "casque audio"},
	{"écouteurs", "écouteurs"},
	{"airpods", "écouteurs"},
	{"enceinte", "enceinte"},
	{"aspirateur", "aspirateur"},
	{"roomba", "aspirateur robot"},
	{"réfrigérateur", "réfrigérateur"},
	{"refrigerateur", "réfrigérateur"},
	{"lave-linge", "lave-linge"},
	{"console", "console de jeu"},
	{"ps5", "console de jeu"},
	{"playstation", "console de jeu"},
	{"xbox", "console de jeu"},
	{"nintendo", "console de jeu"},
	{"appareil photo", "appareil photo"},
	{"camera", "caméra"},
	{"montre", "montre"},
	{"watch", "montre"},
	{"smartwatch", "smartwatch"},
	{"vélo", "vélo"},
	{"velo", "vélo"},
	{"trottinette", "trottinette électrique"},
	{"parfum", "parfum"},
	{"sneakers", "chaussures de sport"},
	{"chaussures", "chaussures"},
}

// warrantyHints maps product types to warranty hints.
var warrantyHints = map[string]string{
	"smartphone":         "2 ans constructeur typique (EU)",
	"laptop":             "2 ans constructeur typique (EU)",
	"tablette":           "2 ans constructeur typique (EU)",
	"télévision":         "2–3 ans constructeur typique",
	"casque audio":       "1–2 ans constructeur typique",
	"écouteurs":          "1 an constructeur typique",
	"enceinte":           "1–2 ans constructeur typique",
	"aspirateur":         "2 ans constructeur typique",
	"aspirateur robot":   "2 ans constructeur typique",
	"réfrigérateur":      "2–5 ans constructeur typique",
	"lave-linge":         "2–5 ans constructeur typique",
	"console de jeu":     "1 an constructeur typique",
	"appareil photo":     "1–2 ans constructeur typique",
	"caméra":             "1–2 ans constructeur typique",
	"montre":             "2 ans constructeur typique",
	"smartwatch":         "1–2 ans constructeur typique",
	"vélo":               "2 ans cadre typique",
	"trottinette électrique": "1–2 ans constructeur typique",
}

// refurbishedProductTypes is the set of product types commonly sold refurbished.
var refurbishedProductTypes = map[string]bool{
	"smartphone":         true,
	"laptop":             true,
	"tablette":           true,
	"télévision":         true,
	"casque audio":       true,
	"écouteurs":          true,
	"console de jeu":     true,
	"appareil photo":     true,
	"caméra":             true,
	"montre":             true,
	"smartwatch":         true,
}

// Tag derives an ItemTags struct from an item name using pure Go logic.
func Tag(itemID, itemName string) ItemTags {
	lower := strings.ToLower(itemName)

	// Detect brand.
	detectedBrand := ""
	for _, b := range brands {
		if strings.Contains(lower, b) {
			// Capitalise first letter of each word.
			words := strings.Fields(b)
			capped := make([]string, len(words))
			for i, w := range words {
				if len(w) > 0 {
					capped[i] = strings.ToUpper(w[:1]) + w[1:]
				}
			}
			detectedBrand = strings.Join(capped, " ")
			break
		}
	}

	// Detect category.
	suggestedCategory := "Autre"
	for _, rule := range categoryRules {
		for _, kw := range rule.keywords {
			if strings.Contains(lower, kw) {
				suggestedCategory = rule.category
				break
			}
		}
		if suggestedCategory != "Autre" {
			break
		}
	}

	// Detect product type.
	productType := ""
	for _, pt := range productTypeKeywords {
		if strings.Contains(lower, pt.keyword) {
			productType = pt.productType
			break
		}
	}

	// Build tags: words with 4+ chars that are not articles/prepositions.
	stopwords := map[string]bool{
		"avec": true, "pour": true, "dans": true, "sous": true,
		"noir": true, "blanc": true, "gris": true,
		"édition": true, "edition": true, "version": true,
		"modèle": true, "modele": true, "neuf": true, "used": true,
	}
	words := strings.FieldsFunc(lower, func(r rune) bool {
		return r == ' ' || r == ',' || r == '-' || r == '(' || r == ')' || r == '/'
	})
	seen := map[string]bool{}
	tags := []string{}
	if detectedBrand != "" {
		tags = append(tags, detectedBrand)
		seen[strings.ToLower(detectedBrand)] = true
	}
	for _, w := range words {
		if len(w) < 4 || stopwords[w] || seen[w] {
			continue
		}
		seen[w] = true
		// Capitalise.
		display := strings.ToUpper(w[:1]) + w[1:]
		tags = append(tags, display)
		if len(tags) >= 8 {
			break
		}
	}

	// IsLuxury: brand-based only.
	isLuxury := luxuryBrands[strings.ToLower(detectedBrand)]

	// IsTech: category-based.
	isTech := techCategories[suggestedCategory]

	// WarrantyHint.
	warrantyHint := ""
	if productType != "" {
		warrantyHint = warrantyHints[productType]
	}
	if warrantyHint == "" {
		warrantyHint = "2 ans minimum légal (EU)"
	}

	// RefurbishedAvailable.
	refurbishedAvailable := refurbishedProductTypes[productType]

	return ItemTags{
		ItemID:               itemID,
		ItemName:             itemName,
		SuggestedCategory:    suggestedCategory,
		Tags:                 tags,
		Brand:                detectedBrand,
		ProductType:          productType,
		IsLuxury:             isLuxury,
		IsTech:               isTech,
		WarrantyHint:         warrantyHint,
		RefurbishedAvailable: refurbishedAvailable,
		GeneratedAt:          time.Now().UTC(),
	}
}
