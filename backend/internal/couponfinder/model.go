package couponfinder

import "time"

// CouponResult aggregates coupon and cashback opportunities for a shopping item.
type CouponResult struct {
	ItemID               string          `json:"itemId"`
	ItemName             string          `json:"itemName"`
	Merchant             string          `json:"merchant"`
	Coupons              []Coupon        `json:"coupons"`
	CashbackOffers       []CashbackOffer `json:"cashbackOffers"`
	TotalPotentialSaving float64         `json:"totalPotentialSaving"`
	GeneratedAt          time.Time       `json:"generatedAt"`
}

// Coupon represents a single promo code opportunity.
type Coupon struct {
	Code        string  `json:"code"`
	Description string  `json:"description"`
	DiscountPct float64 `json:"discountPct"`
	DiscountAmt float64 `json:"discountAmt"`
	Source      string  `json:"source"`
	ExpiresIn   string  `json:"expiresIn"`
	Confidence  string  `json:"confidence"`
}

// CashbackOffer represents a cashback opportunity on a specific platform.
type CashbackOffer struct {
	Platform    string  `json:"platform"`
	CashbackPct float64 `json:"cashbackPct"`
	MinPurchase float64 `json:"minPurchase"`
	Description string  `json:"description"`
}
