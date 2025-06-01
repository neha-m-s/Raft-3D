package models

type PrintJob struct {
	ID           string `json:"id"`
	PrinterID    string `json:"printer_id"`
	FilamentID   string `json:"filament_id"`
	FilePath     string `json:"filepath"`
	PrintWeight  int    `json:"print_weight_in_grams"`
	Status       string `json:"status"` // Queued, Running, Canceled, Done
}
