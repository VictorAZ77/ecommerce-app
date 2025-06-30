package repositories

import (
	"backend/internal/models"
	"database/sql"
)

type UserRepository interface {
	Create(user models.User) (*models.User, error)
	GetByID(id string) (*models.User, error)
	GetByUsername(username string) (*models.User, error)
	GetByEmail(email string) (*models.User, error)
	Update(user models.User) error
	Delete(id string) error
	GetDB() *sql.DB
}
