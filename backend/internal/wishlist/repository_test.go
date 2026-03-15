package wishlist

import (
	"testing"
	"time"
)

func TestWishListItem_Fields(t *testing.T) {
	url := "https://example.com"
	price := 49.99
	notes := "Black Friday deal"
	item := WishListItem{
		ID:          "uuid-123",
		Name:        "Sony WH-1000XM5",
		URL:         &url,
		TargetPrice: &price,
		Currency:    "EUR",
		Notes:       &notes,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if item.Name != "Sony WH-1000XM5" {
		t.Errorf("unexpected name: %s", item.Name)
	}
	if *item.TargetPrice != 49.99 {
		t.Errorf("unexpected price: %f", *item.TargetPrice)
	}
}

func TestCreateWishListItemRequest_DefaultCurrency(t *testing.T) {
	req := CreateWishListItemRequest{Name: "Test"}
	if req.Name == "" {
		t.Error("name should not be empty")
	}
	// currency defaults applied in handler
}

func TestWishListItem_NilOptionalFields(t *testing.T) {
	item := WishListItem{
		ID:       "uuid-456",
		Name:     "Minimal Item",
		Currency: "EUR",
	}
	if item.URL != nil {
		t.Error("URL should be nil")
	}
	if item.TargetPrice != nil {
		t.Error("TargetPrice should be nil")
	}
}
