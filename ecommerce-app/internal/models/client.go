package models

import "time"

//campos del cliente
type Client struct {
	ID        string
	Name      string
	Email     string
	Address   string
	CreatedAt time.Time
	UpdatedAt time.Time
}
