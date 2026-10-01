package api

type Order struct {
	UserID string `json:"user_id"`
	Amount int    `json:"amount"`
}
