package models

import "time"

//Campos del producto
type Product struct {
	ID          string // Identificador único del producto
	Name        string
	Description string
	Price       float64
	Stock       int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
