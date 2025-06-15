package repositories

import (
	"backend/internal/models"
	"database/sql"
)

// UserRepository define los métodos para interactuar con la base de datos para los usuarios.
type UserRepository interface {
	Create(user models.User) (*models.User, error)
	GetByID(id string) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	Update(user models.User) error
	Delete(id string) error
	GetDB() *sql.DB
}
