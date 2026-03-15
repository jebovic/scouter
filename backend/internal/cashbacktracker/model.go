package cashbacktracker

type MerchantCashback struct {
	Merchant        string  `json:"merchant"`
	Platform        string  `json:"platform"` // iGraal, Poulpeo, Cashback World, etc.
	CashbackPct     float64 `json:"cashbackPct"`
	EstimatedAmount float64 `json:"estimatedAmount"` // based on item prices
	ItemCount       int     `json:"itemCount"`
}

type CashbackSummary struct {
	MissionID        string             `json:"missionId"`
	MerchantCashback []MerchantCashback `json:"merchantCashback"`
	TotalEstimated   float64            `json:"totalEstimated"`
	BestPlatform     string             `json:"bestPlatform"`
	Tip              string             `json:"tip"`
}
