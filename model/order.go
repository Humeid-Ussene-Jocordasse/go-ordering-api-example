package model

import (
	"github.com/google/uuid"
	"time"
)

// Remove the json tags before to see what is going to happen
type Order struct {
	// What happens when we don´t use the json tags
	OrderId    uint64     `json:"order_id"`
	CustomerID uuid.UUID  `json:"customer_id"`
	LineItems  []LineItem `json:"line_items"`
	//OrderStatus string
	CreateAt    *time.Time `json:"create_at"`
	ShippedAt   *time.Time `json:"shipped_at"`
	CompletedAt *time.Time `json:"completed_at"`
}

type LineItem struct {
	ItemID   uuid.UUID `json:"item_id"`
	Quantity uint      `json:"quantity"`
	Price    uint      `json:"price"`
}
