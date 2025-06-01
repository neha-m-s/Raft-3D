package models

type Filament struct {
	ID                 string `json:"id"`
	Type               string `json:"type"`
	Color              string `json:"color"`
	TotalWeight        int    `json:"total_weight_in_grams"`
	RemainingWeight    int    `json:"remaining_weight_in_grams"`
}
