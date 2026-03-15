package marketplace

import (
	"fmt"
	"net/url"
)

// Site describes a French e-commerce marketplace and how to build a search URL.
type Site struct {
	Name     string
	SearchFn func(query string) string
	Label    string
}

// Sites is the static list of supported French marketplaces.
var Sites = []Site{
	{
		Name:  "Amazon FR",
		Label: "Neuf",
		SearchFn: func(q string) string {
			return "https://www.amazon.fr/s?k=" + url.QueryEscape(q)
		},
	},
	{
		Name:  "Fnac",
		Label: "Neuf",
		SearchFn: func(q string) string {
			return "https://www.fnac.com/SearchResult/ResultSet.aspx?Search=" + url.QueryEscape(q)
		},
	},
	{
		Name:  "Cdiscount",
		Label: "Neuf",
		SearchFn: func(q string) string {
			return fmt.Sprintf("https://www.cdiscount.com/search/10/%s.html", url.QueryEscape(q))
		},
	},
	{
		Name:  "Leboncoin",
		Label: "Occasion",
		SearchFn: func(q string) string {
			return "https://www.leboncoin.fr/recherche?text=" + url.QueryEscape(q)
		},
	},
	{
		Name:  "Vinted",
		Label: "Seconde main",
		SearchFn: func(q string) string {
			return "https://www.vinted.fr/catalog?search_text=" + url.QueryEscape(q)
		},
	},
	{
		Name:  "Rakuten FR",
		Label: "Neuf/Occasion",
		SearchFn: func(q string) string {
			return "https://fr.shopping.rakuten.com/search?kw=" + url.QueryEscape(q)
		},
	},
	{
		Name:  "Backmarket",
		Label: "Reconditionné",
		SearchFn: func(q string) string {
			return "https://www.backmarket.fr/fr-fr/search?q=" + url.QueryEscape(q)
		},
	},
	{
		Name:  "Boulanger",
		Label: "Neuf",
		SearchFn: func(q string) string {
			return "https://www.boulanger.com/recherche/" + url.QueryEscape(q)
		},
	},
}
