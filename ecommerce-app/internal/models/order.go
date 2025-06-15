package models

import "time"

//campos del Pedido
type Order struct {
	ID          string // Identificador único del pedido
	ClientID    string
	OrderDate   time.Time
	TotalAmount float64
	Status      string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

//campos de la orden del Pedido
type OrderItem struct {
	ID        string // Identificador único del ítem del pedido
	OrderID   string
	ProductID string
	Quantity  int
	Price     float64
	CreatedAt time.Time
	UpdatedAt time.Time
}
