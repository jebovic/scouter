package wishlistvote

// VotedItem holds aggregated vote data for a single option in a mission.
type VotedItem struct {
	ItemID       string  `json:"itemId"`
	Name         string  `json:"name"`
	Price        float64 `json:"price"`
	Ups          int     `json:"ups"`
	Downs        int     `json:"downs"`
	NetScore     int     `json:"netScore"`
	TotalVotes   int     `json:"totalVotes"`
	ApprovalRate float64 `json:"approvalRate"`
	Sentiment    string  `json:"sentiment"`
}

// VoteSummaryResponse is the full response for GET /api/missions/{missionId}/vote-summary.
type VoteSummaryResponse struct {
	MissionID        string      `json:"missionId"`
	Items            []VotedItem `json:"items"`
	MostApproved     *VotedItem  `json:"mostApproved"`
	MostControversial *VotedItem `json:"mostControversial"`
}

// sentiment returns a French sentiment label based on approval rate and total votes.
func sentiment(approvalRate float64, totalVotes int) string {
	if totalVotes == 0 {
		return "aucun vote"
	}
	switch {
	case approvalRate > 0.7:
		return "forte approbation"
	case approvalRate > 0.5:
		return "approuvé"
	case approvalRate >= 0.3:
		return "mitigé"
	default:
		return "désapprouvé"
	}
}
