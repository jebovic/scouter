package budgetadvisor

type AllocationItem struct {
	ItemID          string  `json:"itemId"`
	Name            string  `json:"name"`
	Price           float64 `json:"price"`
	AllocatedBudget float64 `json:"allocatedBudget"` // recommended spend
	Priority        int     `json:"priority"`          // 1=highest
	Reason          string  `json:"reason"`            // French reasoning
	Feasible        bool    `json:"feasible"`          // can afford with remaining budget
}

type BudgetAdvice struct {
	MissionID      string           `json:"missionId"`
	TotalBudget    float64          `json:"totalBudget"`
	Spent          float64          `json:"spent"`
	Remaining      float64          `json:"remaining"`
	Allocations    []AllocationItem `json:"allocations"`
	TotalAllocated float64          `json:"totalAllocated"`
	Surplus        float64          `json:"surplus"` // remaining - totalAllocated
	Advice         string           `json:"advice"`  // French summary
}
